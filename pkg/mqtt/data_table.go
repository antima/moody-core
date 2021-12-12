package mqtt

import (
	"context"
	"sync"
)

// DataObservable provides an interface for updatable and
// observable types, that can notify a number of entities
// of their state updates
type DataObservable interface {
	Notify(ctx context.Context, event StateTuple)
	Attach(obs chan<- StateTuple)
	Detach(obs chan<- StateTuple)
}

// StateTuple used to communicate with services that receive
// data from the MQTT flows
type StateTuple struct {
	topic string
	state string
}

// TopicManager structs handle the data traffic for each
// MQTT topic flow, with respect to every service using the
// managed topic
type TopicManager struct {
	obsMutex   sync.Mutex
	state      string
	observers  []chan<- StateTuple
	cancelFunc context.CancelFunc
}

// DataTable is a synchronized map-like collection that
// keeps references to the active MQTT topics and their
// manager
type DataTable struct {
	rwMutex    sync.RWMutex
	topicTable map[string]*TopicManager
}

// NewDataTable returns an initialized pointer to a DataTable
func NewDataTable() *DataTable {
	return &DataTable{
		topicTable: make(map[string]*TopicManager),
	}
}

// Add the most recently received payload for the passed topic
// to the table. This function initializes the data handler
// for that topic if it was not already initialized
func (table *DataTable) Add(topic string, state string) {
	table.rwMutex.Lock()
	defer table.rwMutex.Unlock()

	manager, isPresent := table.topicTable[topic]
	if !isPresent {
		table.topicTable[topic] = &TopicManager{}
		manager = table.topicTable[topic]
	}

	table.topicTable[topic].state = state

	ctx, cancelFunc := context.WithCancel(context.Background())
	if manager.cancelFunc != nil {
		manager.cancelFunc()
	}

	manager.cancelFunc = cancelFunc
	go manager.Notify(ctx, StateTuple{topic, state})
}

// Get the latest reading for the passed topic, the second return
// value is false if such a topic does not exist in the table
func (table *DataTable) Get(topic string) (string, bool) {
	table.rwMutex.RLock()
	defer table.rwMutex.RUnlock()

	value, isPresent := table.topicTable[topic]
	return value.state, isPresent
}

func (table *DataTable) getManagerRef(topic string) *TopicManager {
	mgr, isPresent := table.topicTable[topic]
	if !isPresent {
		table.topicTable[topic] = &TopicManager{}
		mgr = table.topicTable[topic]
	}
	return mgr
}

func NewTopicManager() *TopicManager {
	return &TopicManager{
		observers:  make([]chan<- StateTuple, 5),
		cancelFunc: nil,
	}
}

func (manager *TopicManager) Notify(ctx context.Context, event StateTuple) {
	for _, obsChan := range manager.observers {
		select {
		case <-ctx.Done():
			return
		default:
			obsChan <- event
		}
	}
}

func (manager *TopicManager) Attach(obs chan<- StateTuple) {
	manager.obsMutex.Lock()
	defer manager.obsMutex.Unlock()
	manager.observers = append(manager.observers, obs)
}

func (manager *TopicManager) Detach(obs chan<- StateTuple) {
	// TODO this may be optimized without re-slicing
	// for example keep it there and when you notify something in the future
	// check if the channel is open, if it is not, delete it from the list
	// a call to this would only close the chan
	manager.obsMutex.Lock()
	defer manager.obsMutex.Unlock()
	for idx, obsChan := range manager.observers {
		if obsChan == obs {
			manager.observers = append(manager.observers[:idx], manager.observers[idx+1:]...)
		}
	}
}

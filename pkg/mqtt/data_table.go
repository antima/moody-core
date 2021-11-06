package mqtt

import (
	"context"
	"sync"
)

type DataObservable interface {
	Notify(ctx context.Context, event StateTuple)
	Attach(obs chan<- StateTuple)
	Detach(obs chan<- StateTuple)
}

type StateTuple struct {
	topic string
	state string
}

type TopicManager struct {
	obsMutex   sync.Mutex
	state      string
	observers  []chan<- StateTuple
	cancelFunc context.CancelFunc
}

type DataTable struct {
	rwMutex    sync.RWMutex
	topicTable map[string]TopicManager
}

func NewDataTable() *DataTable {
	return &DataTable{
		topicTable: make(map[string]TopicManager),
	}
}

func (table *DataTable) Add(topic string, state string) {
	table.rwMutex.Lock()
	defer table.rwMutex.Unlock()

	manager, isPresent := table.topicTable[topic]
	if !isPresent {
		manager = TopicManager{}
	}

	manager.state = state

	ctx, cancelFunc := context.WithCancel(context.Background())
	if manager.cancelFunc != nil {
		manager.cancelFunc()
	}

	manager.cancelFunc = cancelFunc
	go manager.Notify(ctx, StateTuple{topic, state})
}

func (table *DataTable) Get(topic string) (string, bool) {
	table.rwMutex.RLock()
	defer table.rwMutex.RUnlock()

	value, isPresent := table.topicTable[topic]
	return value.state, isPresent
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
	manager.observers = append(manager.observers, obs)
}

func (manager *TopicManager) Detach(obs chan<- StateTuple) {
	// TODO this may be optimized without re-slicing
	manager.obsMutex.Lock()
	defer manager.obsMutex.Unlock()
	for idx, obsChan := range manager.observers {
		if obsChan == obs {
			manager.observers = append(manager.observers[:idx], manager.observers[idx+1:]...)
		}
	}
}

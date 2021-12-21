//go:build linux
// +build linux

package mqtt

import (
	"fmt"
	"plugin"
)

var (
	ErrInvalidNameVar    = fmt.Errorf("the Name variable defined in the service is not valid")
	ErrInvalidVersionVar = fmt.Errorf("the Version variable defined in the service is not valid")
	ErrInvalidTopicsVar  = fmt.Errorf("the Topics array defined in the service is not valid")
	ErrInvalidInitFunc   = fmt.Errorf("the init function defined in the service is not valid")
	ErrActuateInitFunc   = fmt.Errorf("the actuate function defined in the service is not valid")
)

// PluginService represent a kind of plugin that is implemented
// as a go plugin module
type PluginService struct {
	dataChan    chan StateTuple
	Name        string `json:"name"`
	ServiceName string `json:"serviceName"`
	Version     string `json:"version"`
	topics      []string
	init        func() error
	actuate     func(topic string, state string) error
}

// NewPluginService creates a new service from the passed plugin
// file, returning an error if there is no such file or if it does
// not conform to the moody service interface
func NewPluginService(filename string) (*PluginService, error) {
	pluginService, err := plugin.Open(filename)
	if err != nil {
		return nil, err
	}

	name, err := pluginService.Lookup("Name")
	nameVar, isNameVar := name.(*string)
	if err != nil {
		return nil, err
	}

	if !isNameVar {
		return nil, ErrInvalidNameVar
	}

	version, err := pluginService.Lookup("Version")
	versionVar, isVersionVar := version.(*string)
	if err != nil {
		return nil, err
	}

	if !isVersionVar {
		return nil, ErrInvalidVersionVar
	}

	topics, err := pluginService.Lookup("Topics")
	topicsVar, isTopicsVar := topics.(*[]string)
	if err != nil {
		return nil, err
	}

	if !isTopicsVar {
		return nil, ErrInvalidTopicsVar
	}

	init, err := pluginService.Lookup("Init")
	initFunc, isInitFunc := init.(func() error)
	if err != nil {
		return nil, err
	}

	if !isInitFunc {
		return nil, ErrInvalidInitFunc
	}

	actuate, err := pluginService.Lookup("Actuate")
	actuateFunc, isActuateFunc := actuate.(func(string, string) error)
	if err != nil {
		return nil, err
	}

	if !isActuateFunc {
		return nil, ErrActuateInitFunc
	}

	for idx, topic := range *topicsVar {
		(*topicsVar)[idx] = fmt.Sprintf("%s%s", baseTopic[:len(baseTopic)-1], topic)
	}

	return &PluginService{
		dataChan:    make(chan StateTuple),
		Name:        filename,
		ServiceName: *nameVar,
		Version:     *versionVar,
		topics:      *topicsVar,
		init:        initFunc,
		actuate:     actuateFunc,
	}, nil
}

// Init initializes the service by calling the underlying
// init function
func (service *PluginService) Init() error {
	return service.init()
}

// Topics returns a list of the topics that the service is
// subscribed to
func (service *PluginService) Topics() []string {
	return service.topics
}

// Actuate a (topic, state) tuple
func (service *PluginService) Actuate(topic string, state string) error {
	return service.actuate(topic, state)
}

// ListenForUpdates starts the event loop for the service
func (service *PluginService) ListenForUpdates() {
	for data := range service.dataChan {
		service.actuate(data.topic, data.state)
	}
}

// Stop terminates the service
func (service *PluginService) Stop(dataTable *DataTable) {
	for _, topic := range service.Topics() {
		topicManager := dataTable.getManagerRef(topic)
		topicManager.Detach(service.dataChan)
	}
	close(service.dataChan)
}

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

type PluginService struct {
	dataChan    chan StateTuple
	Name        string `json:"name"`
	ServiceName string `json:"serviceName"`
	Version     string `json:"version"`
	topics      []string
	init        func() error
	actuate     func(topic string, state string) error
}

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

func (service *PluginService) Init() error {
	return service.init()
}

func (service *PluginService) Topics() []string {
	return service.topics
}

func (service *PluginService) Actuate(topic string, state string) error {
	return service.actuate(topic, state)
}

func (service *PluginService) ListenForUpdates() {
	for data := range service.dataChan {
		service.actuate(data.topic, data.state)
	}
}

func (service *PluginService) Stop(dataTable *DataTable) {
	for _, topic := range service.Topics() {
		topicManager := dataTable.getManagerRef(topic)
		topicManager.Detach(service.dataChan)
	}
	close(service.dataChan)
}

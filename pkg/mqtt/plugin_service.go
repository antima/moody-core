package mqtt

import (
	"fmt"
	"plugin"
)

var (
	ErrInvalidInitFunc = fmt.Errorf("the init function defined in the service is not valid")
	ErrActuateInitFunc = fmt.Errorf("the actuate function defined in the service is not valid")
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
	if err != nil {
		return nil, err
	}

	version, err := pluginService.Lookup("Version")
	if err != nil {
		return nil, err
	}

	topics, err := pluginService.Lookup("Topics")
	if err != nil {
		return nil, err
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

	return &PluginService{
		Name:        filename,
		ServiceName: *name.(*string),
		Version:     *version.(*string),
		topics:      *topics.(*[]string),
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

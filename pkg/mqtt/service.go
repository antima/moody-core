package mqtt

import (
	"io/fs"
	"path/filepath"
	"plugin"
)

type InitFunc func() error
type ActuateFunc func(tuple StateTuple) error

type MoodyService interface {
	Name() string
	ServiceName() string
	Version() string
	Init() error
	Actuate(PublishFunc) error
}

type PluginService struct {
	dataChan <-chan StateTuple

	name        string
	serviceName string
	version     string
	topics      []string
	init        InitFunc
	actuate     ActuateFunc
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
	if err != nil {
		return nil, err
	}

	actuate, err := pluginService.Lookup("Actuate")
	if err != nil {
		return nil, err
	}

	return &PluginService{
		name:        filename,
		serviceName: *name.(*string),
		version:     *version.(*string),
		topics:      *topics.(*[]string),
		init:        init.(InitFunc),
		actuate:     actuate.(ActuateFunc),
	}, nil
}

func (service *PluginService) Name() string {
	return service.name
}

func (service *PluginService) Version() string {
	return service.version
}

func (service *PluginService) ListenForUpdates() {
	for data := range service.dataChan {
		service.actuate(data)
	}
}

func GetAllServices(serviceDir string) ([]string, error) {
	var services []string
	err := filepath.WalkDir(serviceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".so" {
			services = append(services, d.Name())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}

func StartupService(serviceFile string) {
}

func ServiceFileWatch() {
}

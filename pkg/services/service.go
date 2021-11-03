package services

import (
	"io/fs"
	"path/filepath"
	"plugin"
)

// load at startup via conf in future
const serviceFolder = "~/moody-services"


type InitFunc func() error
type ActuateFunc func() error

type MoodyService interface {
	Name() string
	Version() string
	Init() error
	Actuate() error
}

type PluginService struct {
	name string
	version string
	init InitFunc
	actuate ActuateFunc
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

	init, err := pluginService.Lookup("Init")
	if err != nil {
		return nil, err
	}

	actuate, err := pluginService.Lookup("Actuate")
	if err != nil {
		return nil, err
	}

	return &PluginService{
		name: *name.(*string),
		version: *version.(*string),
		init: init.(InitFunc),
		actuate: actuate.(ActuateFunc),
	}, nil
}

func (service *PluginService) Name() string {
	return service.name
}

func (service *PluginService) Version() string {
	return service.version
}

func (service *PluginService) Init() error {
	return service.init()
}

func (service *PluginService) Actuate() error {
	return service.actuate()
}

func GetAllServices(serviceDir string) ([]string, error) {
	var services []string
	err := filepath.WalkDir(serviceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".so" {
			services= append(services, d.Name())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}


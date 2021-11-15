package mqtt

import (
	"io/fs"
	"path/filepath"
	"plugin"
	"time"
)

type InitFunc func() error
type ActuateFunc func(tuple StateTuple) error

var (
	services *ConcurrentSet = NewConcurrentSet()
)

type MoodyService interface {
	Name() string
	ServiceName() string
	Version() string
}

type PluginService struct {
	dataChan    chan StateTuple
	Name        string `json:"name"`
	ServiceName string `json:"serviceName"`
	Version     string `json:"version"`
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
		Name:        filename,
		ServiceName: *name.(*string),
		Version:     *version.(*string),
		topics:      *topics.(*[]string),
		init:        init.(InitFunc),
		actuate:     actuate.(ActuateFunc),
	}, nil
}

func (service *PluginService) ListenForUpdates() {
	for data := range service.dataChan {
		service.actuate(data)
	}
}

func StartServiceManager(serviceDir string, dataTable *DataTable) {
	serviceNames := getAllServices(serviceDir)
	startupServices(serviceNames, dataTable)

	for {
		select {
		case <-time.After(1 * time.Second):
			currServiceNames := getAllServices(serviceDir)
			toAdd := serviceNames.Difference(currServiceNames)
			toDel := currServiceNames.Difference(serviceNames)

			if toAdd.Size() > 0 {
				startupServices(toAdd, dataTable)

			}
			if toAdd.Size() > 0 {
				stopServices(toDel, dataTable)
			}
		}
	}
}

func GetActiveServices() []*PluginService {
	return nil
}

func startupServices(serviceNames *ConcurrentSet, dataTable *DataTable) {
	if serviceNames == nil || serviceNames.Size() == 0 {
		return
	}
	servIter := serviceNames.Iterator()
	for next, end := servIter.Next(); !end; next, end = servIter.Next() {
		service, err := NewPluginService(next.(string))
		if err == nil {
			services.Add(service)
			service.init()
			for _, topic := range service.topics {
				mgr := dataTable.getManagerRef(topic)
				mgr.Attach(service.dataChan)
			}
			go service.ListenForUpdates()
		}
	}
}

func stopServices(serviceNames *ConcurrentSet, dataTable *DataTable) {
	if serviceNames == nil || serviceNames.Size() == 0 {
		return
	}
	servIter := serviceNames.Iterator()
	for next, end := servIter.Next(); !end; next, end = servIter.Next() {
		services.Remove(next)
	}
}

func getAllServices(serviceDir string) *ConcurrentSet {
	serviceNames := NewConcurrentSet()
	_ = filepath.WalkDir(serviceDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(d.Name()) == ".so" {
			serviceNames.Add(d.Name())
		}
		return nil
	})
	return serviceNames
}

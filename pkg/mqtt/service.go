package mqtt

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"time"
)

var (
	// may change this into a goroutine able to return copies
	// from a channel for the public HTTP APIs
	services = NewServiceMap()
)

type MoodyService interface {
	Init() error
	Topics() []string
	Actuate(topic string, state string) error
	ListenForUpdates()
	Stop(dataTable *DataTable)
}

func StartServiceManager(serviceDir string, dataTable *DataTable) {
	log.Printf("Starting the service manager module, serving services from %s\n", serviceDir)
	go func() {
		serviceNames := getAllServices(serviceDir)
		if serviceNames.Size() > 0 {
			startupServices(serviceNames, dataTable)
		}

		for {
			select {
			case <-time.After(1 * time.Second):
				currServiceNames := getAllServices(serviceDir)
				toAdd := currServiceNames.Difference(serviceNames)
				toDel := serviceNames.Difference(currServiceNames)

				if toAdd.Size() > 0 {
					startupServices(toAdd, dataTable)
					iter := toAdd.Iterator()
					for next, end := iter.Next(); !end; next, end = iter.Next() {
						serviceNames.Add(next)
					}
				}
				if toDel.Size() > 0 {
					stopServices(toDel, dataTable)
					iter := toAdd.Iterator()
					for next, end := iter.Next(); !end; next, end = iter.Next() {
						serviceNames.Remove(next)
					}
				}
			}
		}
	}()
}

func GetActiveServices() []MoodyService {
	return services.List()
}

func startupServices(serviceNames *ConcurrentSet, dataTable *DataTable) {
	if serviceNames == nil || serviceNames.Size() == 0 {
		return
	}
	servIter := serviceNames.Iterator()
	for next, end := servIter.Next(); !end; next, end = servIter.Next() {
		serviceName := next.(string)
		service, err := NewPluginService(serviceName)
		if err != nil {
			log.Printf("error: could not initialize service '%s', %v", serviceName, err)
			continue
		}

		log.Printf("found service %s\n", service.ServiceName)
		services.Add(next.(string), service)

		err = service.Init()
		if err != nil {
			log.Printf("error: could not initialize service '%s', %v", service.ServiceName, err)
			return
		}

		for _, topic := range service.topics {
			mgr := dataTable.getManagerRef(topic)
			mgr.Attach(service.dataChan)
		}
		log.Printf("service %s starting\n", service.ServiceName)
		go service.ListenForUpdates()

	}
}

func stopServices(serviceNames *ConcurrentSet, dataTable *DataTable) {
	if serviceNames == nil || serviceNames.Size() == 0 {
		return
	}
	servIter := serviceNames.Iterator()
	for next, end := servIter.Next(); !end; next, end = servIter.Next() {
		serviceName := next.(string)
		service, isContained := services.Get(serviceName)
		if isContained {
			log.Printf("service %s stopping\n", serviceName)
			service.Stop(dataTable)
			services.Remove(serviceName)
		}
	}
}

func getAllServices(serviceDir string) *ConcurrentSet {
	serviceNames := NewConcurrentSet()
	_ = filepath.WalkDir(serviceDir, func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.IsDir() && filepath.Ext(d.Name()) == ".so" {
			currName := fmt.Sprintf("%s/%s", serviceDir, d.Name())
			name, err := filepath.Abs(currName)
			if err == nil {
				serviceNames.Add(name)
			}
		}
		return nil
	})
	return serviceNames
}

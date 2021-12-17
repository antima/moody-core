package mqtt

import "sync"

// ServiceMap implements a synchronized collection of services
type ServiceMap struct {
	mutex    sync.RWMutex
	mappings map[string]MoodyService
}

// NewServiceMap creates a new initialized map and returns a pointer to it
func NewServiceMap() *ServiceMap {
	return &ServiceMap{
		mappings: make(map[string]MoodyService),
	}
}

// Add a service to the map, identified by name, in a synchronous fashion
func (concurrentMap *ServiceMap) Add(name string, service MoodyService) {
	concurrentMap.mutex.Lock()
	defer concurrentMap.mutex.Unlock()
	concurrentMap.mappings[name] = service
}

// Get a service from the map in a synchronous fashion, returns (nil, false)
// if the name-key is not present
func (concurrentMap *ServiceMap) Get(name string) (MoodyService, bool) {
	concurrentMap.mutex.RLock()
	defer concurrentMap.mutex.RUnlock()
	elem, isPresent := concurrentMap.mappings[name]
	return elem, isPresent
}

// Remove a service identified by name from the map, returns (nil, false) if
// there is no such element
func (concurrentMap *ServiceMap) Remove(name string) (MoodyService, bool) {
	concurrentMap.mutex.Lock()
	defer concurrentMap.mutex.Unlock()
	elem, isPresent := concurrentMap.mappings[name]
	delete(concurrentMap.mappings, name)
	return elem, isPresent
}

// List returns a slice of all the services in the map
func (concurrentMap *ServiceMap) List() []MoodyService {
	concurrentMap.mutex.RLock()
	defer concurrentMap.mutex.RUnlock()
	serviceList := make([]MoodyService, len(concurrentMap.mappings))
	for _, service := range concurrentMap.mappings {
		serviceList = append(serviceList, service)
	}
	return serviceList
}

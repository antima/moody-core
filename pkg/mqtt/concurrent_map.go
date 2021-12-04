package mqtt

import "sync"

type ServiceMap struct {
	mutex    sync.RWMutex
	mappings map[string]MoodyService
}

func NewServiceMap() *ServiceMap {
	return &ServiceMap{
		mappings: make(map[string]MoodyService),
	}
}

func (concurrentMap *ServiceMap) Add(name string, service MoodyService) {
	concurrentMap.mutex.Lock()
	defer concurrentMap.mutex.Unlock()
	concurrentMap.mappings[name] = service
}

func (concurrentMap *ServiceMap) Get(name string) (MoodyService, bool) {
	concurrentMap.mutex.RLock()
	defer concurrentMap.mutex.RUnlock()
	elem, isPresent := concurrentMap.mappings[name]
	return elem, isPresent
}

func (concurrentMap *ServiceMap) Remove(name string) (MoodyService, bool) {
	concurrentMap.mutex.Lock()
	defer concurrentMap.mutex.Unlock()
	elem, isPresent := concurrentMap.mappings[name]
	delete(concurrentMap.mappings, name)
	return elem, isPresent
}

func (concurrentMap *ServiceMap) List() []MoodyService {
	concurrentMap.mutex.RLock()
	defer concurrentMap.mutex.RUnlock()
	serviceList := make([]MoodyService, len(concurrentMap.mappings))
	for _, service := range concurrentMap.mappings {
		serviceList = append(serviceList, service)
	}
	return serviceList
}

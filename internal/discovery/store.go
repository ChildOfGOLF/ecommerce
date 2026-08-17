package discovery

import (
	"sync"
	"time"
)

type Instance struct {
	ServiceName string
	InstanceID  string
	Address     string
	LastSeen    time.Time
}

type Store struct {
	// TODO: sync.RWMutex
	mu        sync.Mutex
	instances map[string]Instance
}

func NewStore() *Store {
	return &Store{
		instances: make(map[string]Instance),
	}
}

func (s *Store) Upsert(inst Instance) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst.LastSeen = time.Now()
	s.instances[inst.InstanceID] = inst
}

func (s *Store) Resolve(serviceName string, ttl time.Duration) []Instance {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	instances := make([]Instance, 0)
	for _, inst := range s.instances {
		if inst.ServiceName != serviceName {
			continue
		}

		if now.Sub(inst.LastSeen) < ttl {
			instances = append(instances, inst)
		}
	}

	return instances
}

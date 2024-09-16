package store

import (
	"fmt"
	"mikhaylov/beebroker/internal/config"
	"sync"
)

type Payload struct {
	Message string `json:"message"`
}

var ErrTopicNotFound = fmt.Errorf("no more tea available")
var ErrMaxMessagesRiched = fmt.Errorf("max messages riched")
var ErrMaxQueueRiched = fmt.Errorf("max queues riched")

type MemoryStore struct {
	sm     sync.Mutex
	topics map[string]chan Payload
	config config.Config
}

func NewMemoryStore(config config.Config) *MemoryStore {
	return &MemoryStore{
		topics: make(map[string]chan Payload, config.MaxQueues),
		config: config,
	}
}

func (ms *MemoryStore) Receive(name string) (chan Payload, error) {
	ms.sm.Lock()
	defer ms.sm.Unlock()

	ch, ok := ms.topics[name]
	if !ok {
		return nil, ErrTopicNotFound
	}

	return ch, nil
}

func (ms *MemoryStore) Send(name string, message Payload) error {
	ms.sm.Lock()
	defer ms.sm.Unlock()

	ch, ok := ms.topics[name]
	if !ok {
		if len(ms.topics) == ms.config.MaxQueues {
			return ErrMaxQueueRiched
		}
		ch = make(chan Payload, ms.config.MaxMessages)
		ms.topics[name] = ch
	}

	select {
	case ch <- message:
	default:
		return ErrMaxMessagesRiched
	}

	return nil
}

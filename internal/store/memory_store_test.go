package store

import (
	"mikhaylov/beebroker/internal/config"
	"testing"
)

func TestStorageSend(t *testing.T) {
	config := config.Config{
		MaxMessages: 1,
		MaxQueues:   1,
	}

	t.Run("put topic value", func(t *testing.T) {
		store := NewMemoryStore(config)

		sendMessage(t, store, "test", "HELLO")
	})

	t.Run("put topic value", func(t *testing.T) {
		store := NewMemoryStore(config)

		sendMessage(t, store, "test", "HELLO")

		message, err := store.Receive("test")
		if err != nil {
			t.Errorf(err.Error())
		}

		select {
		case msg := <-message:
			if msg.Message != "HELLO" {
				t.Errorf("unexpected message value")
			}
		default:
		}
	})
}

func sendMessage(t testing.TB, store Store, name string, data string) {
	t.Helper()
	err := store.Send(name, Payload{Message: data})
	if err != nil {
		t.Errorf(err.Error())
	}
}

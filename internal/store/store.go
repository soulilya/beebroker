package store

type Store interface {
	Receive(name string) (chan Payload, error)
	Send(name string, message Payload) error
}

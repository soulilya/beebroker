package config

type Config struct {
	Host           string
	Port           int
	Timeout        int
	MaxMessages    int
	MaxQueues      int
	MaxConnections int
	MaxRequests    int
}

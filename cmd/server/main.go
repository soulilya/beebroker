package main

import (
	"context"
	"flag"
	"fmt"
	"mikhaylov/beebroker/internal/config"
	"mikhaylov/beebroker/internal/http"
	"mikhaylov/beebroker/internal/store"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const ShutdownTimeout = 1 * time.Second

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	Run()
}

func Run() error {
	host := flag.String("host", "localhost", "Host must be ip or domain")
	port := flag.Int("port", 8080, "Port must be int value")
	timeout := flag.Int("timeot", 60, "Timeout connection in seconds. Recommended value between 60-120 seconds.")
	maxMessages := flag.Int("max_messages", 1000, "Maximum messages in queue")
	maxQueue := flag.Int("max_queue", 1000, "Maximum queue count")
	maxConn := flag.Int("max_connections", 20, "Maximum connections count")

	flag.Parse()

	config := config.Config{
		Host:           *host,
		Port:           *port,
		Timeout:        *timeout,
		MaxMessages:    *maxMessages,
		MaxQueues:      *maxQueue,
		MaxConnections: *maxConn,
	}

	store := store.NewMemoryStore(config)
	srv := http.NewBrokerServer(config, store)

	if err := srv.Open(); err != nil {
		fmt.Printf("HTTP shutdown error: %v\n", err)
	}

	fmt.Printf("Server started on %s:%d\n", config.Host, config.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Close(shutdownCtx); err != nil {
		fmt.Printf("HTTP shutdown error: %v\n", err)
	}
	fmt.Println("Server stopped")

	return nil
}

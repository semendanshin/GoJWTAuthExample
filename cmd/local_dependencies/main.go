package main

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()

	container, err := mongodb.RunContainer(
		ctx,
		testcontainers.WithImage("mongo:latest"),
		mongodb.WithUsername("root"),
		mongodb.WithPassword("root"),
	)
	if err != nil {
		log.Fatal(err)
	}

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	log.Printf("MongoDB running at: %s", endpoint)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	log.Printf("Shutting down MongoDB container")

	if err := container.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}

}

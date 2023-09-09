package main

import (
	"google.golang.org/grpc"
	"log"
	"sync"
	pb "therealbroker/api/proto/broker/api/proto"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBrokerClient(conn)

	var wg sync.WaitGroup
	ticker := time.NewTicker(50 * time.Microsecond) // 0.5 billion request in 20 minutes

	done := make(chan bool)

	wg.Add(1)
	go func(client pb.BrokerClient) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				go publish(err, client)
			}
		}
	}(client)

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(60 * time.Second)
		ticker.Stop()
		done <- true
	}()
	wg.Wait()
}

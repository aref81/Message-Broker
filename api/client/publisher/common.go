package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	pb "therealbroker/api/proto/broker/api/proto"
	"time"
)

func publish(client pb.BrokerClient, ctx context.Context) error {
	publishResponse, err := client.Publish(ctx, &pb.PublishRequest{
		Subject:           "sample",
		Body:              []byte(fmt.Sprintf("message sent at : %v", time.Now())),
		ExpirationSeconds: 10000,
	})
	if err != nil {
		return err
	}

	logrus.Printf("Published message with ID: %d\n", publishResponse.Id)
	return nil
}

func main() {
	runMass("localhost:8081")
}

func difLoad() {
	var wg sync.WaitGroup
	ticker := time.NewTicker(2000 * time.Microsecond)

	done := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				go func() {
					runSingle("127.0.0.1:59657")
				}()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Minute)
		ticker.Stop()
		done <- true
	}()
	wg.Wait()
}

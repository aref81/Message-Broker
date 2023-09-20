package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"sync"
	pb "therealbroker/api/proto/broker/api/proto"
	"time"
)

func runMass(host string) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.Out = os.Stdout

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("Failed to connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Fatalf("Failed to close: %v", err)
		}
	}(conn)

	client := pb.NewBrokerClient(conn)
	ctx := context.Background()

	var wg sync.WaitGroup
	ticker := time.NewTicker(50 * time.Microsecond)

	done := make(chan bool)

	wg.Add(1)
	go func(client pb.BrokerClient) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				go func() {
					err := publish(client, ctx)
					if err != nil {
						logger.Fatalf("Publish failed: %v", err)
					}
				}()
			}
		}
	}(client)

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Minute)
		ticker.Stop()
		done <- true
	}()
	wg.Wait()
}

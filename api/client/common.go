package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
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

func subscribe(err error, client pb.BrokerClient) pb.Broker_SubscribeClient {
	subscribeRequest := &pb.SubscribeRequest{Subject: "example"}
	stream, err := client.Subscribe(context.Background(), subscribeRequest)
	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}
	return stream
}

func main() {
	runMass("localhost:10000")
}

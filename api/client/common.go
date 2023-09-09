package main

import (
	"context"
	"fmt"
	"log"
	pb "therealbroker/api/proto/broker/api/proto"
	"time"
)

func publish(err error, client pb.BrokerClient) {
	publishResponse, err := client.Publish(context.Background(), &pb.PublishRequest{
		Subject:           "example",
		Body:              []byte(fmt.Sprintf("message sent at : %v", time.Now())),
		ExpirationSeconds: 60,
	})
	if err != nil {
		log.Fatalf("Publish failed: %v", err)
	}
	fmt.Printf("Published message with ID: %d\n", publishResponse.Id)
}

func subscribe(err error, client pb.BrokerClient) pb.Broker_SubscribeClient {
	subscribeRequest := &pb.SubscribeRequest{Subject: "example"}
	stream, err := client.Subscribe(context.Background(), subscribeRequest)
	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}
	return stream
}

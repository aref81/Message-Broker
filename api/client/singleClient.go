package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	pb "therealbroker/api/proto/broker/api/proto"
)

func runSingle(host string) {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBrokerClient(conn)
	err = publish(client, context.Background())

	if err != nil {
		fmt.Println("Publish failed: %v", err)
	}

}

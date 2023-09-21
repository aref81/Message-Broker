package main

import (
	"context"
	"fmt"
	"math/rand"
	"therealbroker/api/proto/broker/api/proto"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

type myBrokerServer struct {
	proto.UnimplementedBrokerServer
	broker broker.Broker
}

func (b myBrokerServer) Publish(ctx context.Context, pubReq *proto.PublishRequest) (*proto.PublishResponse, error) {
	msg := model.Message{
		Id:         0,
		Body:       string(pubReq.Body),
		Expiration: time.Duration(pubReq.ExpirationSeconds),
	}
	id, err := b.broker.Publish(ctx, pubReq.Subject, msg)
	if err != nil {
		return nil, err
	}

	println(id)

	return &proto.PublishResponse{Id: int32(id)}, nil
}

func (b myBrokerServer) Subscribe(subReq *proto.SubscribeRequest, subServer proto.Broker_SubscribeServer) error {
	ch, err := b.broker.Subscribe(subServer.Context(), subReq.Subject)
	if err != nil {
		return err
	}

	for {
		select {
		case msg := <-ch:
			protoMsg := &proto.MessageResponse{Body: []byte(msg.Body)}

			if err := subServer.Send(protoMsg); err != nil {
				return err
			}
		case <-subServer.Context().Done():
			return nil
		}
	}
}

func (b myBrokerServer) Fetch(ctx context.Context, fetchReq *proto.FetchRequest) (*proto.MessageResponse, error) {
	msg, err := b.broker.Fetch(ctx, fetchReq.Subject, int(fetchReq.Id))
	if err != nil {
		return nil, err
	}

	return &proto.MessageResponse{Body: []byte(msg.Body)}, nil
}

func main() {
	//listener, err := net.Listen("tcp", ":8081")
	//
	//if err != nil {
	//	log.Fatal("cannot create listener")
	//}
	//
	//server := grpc.NewServer()
	//proto.RegisterBrokerServer(server, &myBrokerServer{
	//	broker: broker2.NewModule(db.),
	//})
	//
	//if err := server.Serve(listener); err != nil {
	//	log.Fatalf("Server serve failed: %v", err)
	//}

	seen := make(map[int64]bool)

	for i := 0; i < 1000000000; i++ {
		id := generateUniqueID()
		if seen[id] {
			fmt.Println("Duplicates")
			return // Found a duplicate
		}
		seen[id] = true

		if i%10000 == 0 {
			fmt.Println(i)
		}
	}

	fmt.Println("No Duplicates")

}

func generateUniqueID() int64 {
	timestamp := time.Now().UnixNano()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomPart := r.Int63()
	uniqueID := timestamp + randomPart
	if uniqueID < 0 {
		uniqueID = -uniqueID
	}
	return uniqueID
}

func hasDuplicates(arr []int) bool {
	seen := make(map[int]bool)

	for _, item := range arr {
		if seen[item] {
			return true // Found a duplicate
		}
		seen[item] = true
	}

	return false // No duplicates found
}

package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"therealbroker/api/proto/broker/api/proto"
	broker2 "therealbroker/internal/broker"
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
	listener, err := net.Listen("tcp", ":8081")

	if err != nil {
		log.Fatal("cannot create listener")
	}

	server := grpc.NewServer()
	proto.RegisterBrokerServer(server, &myBrokerServer{
		broker: broker2.NewModule(),
	})

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Server serve failed: %v", err)
	}

}

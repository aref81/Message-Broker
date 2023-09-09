package main

//
//import (
//	"fmt"
//	"google.golang.org/grpc"
//	"log"
//	pb "therealbroker/api/proto/broker/api/proto"
//)
//
//func main() {
//	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())
//	if err != nil {
//		log.Fatalf("Failed to connect: %v", err)
//	}
//	defer conn.Close()
//
//	client := pb.NewBrokerClient(conn)
//	stream := subscribe(err, client)
//	publish(err, client)
//
//	for {
//		messageResponse, err := stream.Recv()
//		if err != nil {
//			log.Fatalf("Error receiving message: %v", err)
//			break
//		}
//		fmt.Printf("Received message: %s\n", messageResponse.Body)
//	}
//}

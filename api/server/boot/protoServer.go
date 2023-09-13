package boot

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"therealbroker/api/proto/broker/api/proto"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

type ProtoServer struct {
	proto.UnimplementedBrokerServer
	Broker broker.Broker
	Tracer opentracing.Tracer
}

func (b ProtoServer) Publish(ctx context.Context, pubReq *proto.PublishRequest) (*proto.PublishResponse, error) {
	span := b.Tracer.StartSpan("publish")
	defer span.Finish()
	//startTime := time.Now()
	msg := model.Message{
		Id:         0,
		Body:       string(pubReq.Body),
		Expiration: time.Duration(pubReq.ExpirationSeconds),
	}

	id, err := b.Broker.Publish(ctx, pubReq.Subject, msg)
	if err != nil {
		logRPCError("Publish", err)
		return nil, err
	}

	//observeRPCCall("Publish", startTime)
	successfulRPCCalls.WithLabelValues("Publish").Inc()
	//logrus.Println(fmt.Sprintf("Published Message With ID: %d into Subject: %s", id, pubReq.Subject))
	return &proto.PublishResponse{Id: int32(id)}, nil
}

func (b ProtoServer) Subscribe(subReq *proto.SubscribeRequest, subServer proto.Broker_SubscribeServer) error {
	//startTime := time.Now()

	ch, err := b.Broker.Subscribe(subServer.Context(), subReq.Subject)
	if err != nil {
		logRPCError("Subscribe", err)
		return err
	}

	go func() {
		for {
			select {
			case msg := <-ch:
				protoMsg := &proto.MessageResponse{Body: []byte(msg.Body)}

				if err := subServer.Send(protoMsg); err != nil {
					logRPCError("Subscribe", err)
					return
				}
			case <-subServer.Context().Done():
				//observeRPCCall("Subscribe", startTime)
				return
			}
		}
	}()

	return nil
}

func (b ProtoServer) Fetch(ctx context.Context, fetchReq *proto.FetchRequest) (*proto.MessageResponse, error) {
	//startTime := time.Now()

	msg, err := b.Broker.Fetch(ctx, fetchReq.Subject, int(fetchReq.Id))
	if err != nil {
		logRPCError("Fetch", err)
		return nil, err
	}

	//observeRPCCall("Fetch", startTime)
	return &proto.MessageResponse{Body: []byte(msg.Body)}, nil
}

func logRPCError(method string, err error) {
	statusCode := status.Code(err)
	if statusCode == codes.OK {
		return
	}

	logrus.WithFields(logrus.Fields{
		"method":  method,
		"code":    statusCode,
		"message": err.Error(),
	}).Error("gRPC call failed")
}

func observeRPCCall(method string, startTime time.Time, err error) {
	//elapsed := time.Since(startTime).Seconds()
	//methodDuration.WithLabelValues(method).Observe(elapsed)

	if status.Code(err) == codes.OK {
		successfulRPCCalls.WithLabelValues(method).Inc()
	} else {
		//failedRPCCalls.WithLabelValues(method).Inc()
	}
}
package boot

import (
	"context"
	"fmt"
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
	//span := b.Tracer.StartSpan("publish : protoServer")
	//defer span.Finish()
	startTime := time.Now()

	msg := model.Message{
		Id:         0,
		Body:       string(pubReq.Body),
		Expiration: time.Duration(pubReq.ExpirationSeconds),
	}

	//spandex := opentracing.ContextWithSpan(context.Background(), span)
	id, err := b.Broker.Publish(ctx, pubReq.Subject, msg)

	//Metric
	methodDuration.WithLabelValues("publish").Observe(float64(time.Since(startTime)) / float64(time.Nanosecond))

	if err != nil {
		//Metric
		failedRPCCalls.WithLabelValues("subscribe").Inc()

		logRPCError("Publish", err)
		return nil, err
	}

	//Metric
	successfulRPCCalls.WithLabelValues("Publish").Inc()

	logrus.Println(fmt.Sprintf("Published Message With ID: %d into Subject: %s", id, pubReq.Subject))
	return &proto.PublishResponse{Id: int32(id)}, nil
}

func (b ProtoServer) Subscribe(subReq *proto.SubscribeRequest, subServer proto.Broker_SubscribeServer) error {
	span := b.Tracer.StartSpan("subscribe : protoServer")
	defer span.Finish()
	startTime := time.Now()

	spandex := opentracing.ContextWithSpan(context.Background(), span)
	ch, err := b.Broker.Subscribe(spandex, subReq.Subject)

	//Metric
	methodDuration.WithLabelValues("subscribe").Observe(float64(time.Since(startTime)) / float64(time.Nanosecond))
	activeSubscribersGauge.Inc()

	if err != nil {
		//Metric
		failedRPCCalls.WithLabelValues("subscribe").Inc()
		logRPCError("Subscribe", err)
		return err
	}

	//Metric
	successfulRPCCalls.WithLabelValues("Publish").Inc()

	for {
		select {
		case msg := <-ch:
			protoMsg := &proto.MessageResponse{Body: []byte(msg.Body)}

			if err := subServer.Send(protoMsg); err != nil {
				logRPCError("Subscribe", err)
				activeSubscribersGauge.Dec()
				return err
			}
		case <-subServer.Context().Done():
			logrus.Println("Subscriber Left")
			activeSubscribersGauge.Dec()
			return err
		}
	}
}

//func (b ProtoServer) Subscribe(subReq *proto.SubscribeRequest, subServer proto.Broker_SubscribeServer) error {
//	logrus.Println("REQ")
//
//	ch, _ := b.Broker.Subscribe(context.Background(), subReq.Subject)
//	wg := sync.WaitGroup{}
//
//	wg.Add(1)
//	go func() {
//		for {
//			msg := <-ch
//			protoMsg := &proto.MessageResponse{Body: []byte(msg.Body)}
//
//			if err := subServer.Send(protoMsg); err != nil {
//				logRPCError("Subscribe", err)
//				return
//			}
//		}
//	}()
//
//	wg.Wait()
//	return nil
//}

func (b ProtoServer) Fetch(ctx context.Context, fetchReq *proto.FetchRequest) (*proto.MessageResponse, error) {
	span := b.Tracer.StartSpan("fetch : protoServer")
	defer span.Finish()
	startTime := time.Now()

	spandex := opentracing.ContextWithSpan(context.Background(), span)
	msg, err := b.Broker.Fetch(spandex, fetchReq.Subject, int(fetchReq.Id))

	//Metric
	methodDuration.WithLabelValues("fetch").Observe(float64(time.Since(startTime)) / float64(time.Nanosecond))

	if err != nil {
		//Metric
		failedRPCCalls.WithLabelValues("fetch").Inc()

		logRPCError("Fetch", err)
		return nil, err
	}

	//Metric
	successfulRPCCalls.WithLabelValues("fetch").Inc()

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

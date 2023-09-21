package main

import (
	_ "fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"net"
	"os"
	_ "strconv"
	"therealbroker/api/server/boot"
	"therealbroker/pkg/utils/db"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"therealbroker/api/proto/broker/api/proto"
	broker2 "therealbroker/internal/broker"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.Out = os.Stdout

	// tcp server
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Fatalf("cannot create listener: %v", err)
	}

	//jaeger
	cfg := &config.Configuration{
		ServiceName: "publisher",

		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "jaeger:6831",
		},
	}

	// tracer
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		logger.Fatalf("ERROR: cannot init Jaeger: %v\n", err)
	}
	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			logger.Fatalf("ERROR: cannot close tracer: %v\n", err)
		}
	}(closer)
	opentracing.SetGlobalTracer(tracer)

	// grpc server
	server := grpc.NewServer()
	proto.RegisterBrokerServer(server, &boot.ProtoServer{
		Broker: broker2.NewModule(db.POSTGRES),
		Tracer: nil,
	})

	//prometheus
	go boot.InitPrometheus()

	logger.Println("SERVER STARTED SUCCESSFULLY!")

	if err := server.Serve(listener); err != nil {
		logger.Fatalf("Server serve failed: %v", err)
	}
}

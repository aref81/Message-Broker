package main

import (
	"fmt"
	_ "fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	_ "strconv"
	"therealbroker/api/proto/broker/api/proto"
	"therealbroker/api/server/boot"
	broker2 "therealbroker/internal/broker"
	"therealbroker/pkg/utils/db"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.Out = os.Stdout

	connStr, ex := os.LookupEnv("DBMS")
	if !ex {
		logrus.Printf("The env variable %s is not set.\n", "DBMS")
		connStr = "2"
	}
	fmt.Println(connStr)

	//dbChoice, err := strconv.Atoi(connStr)
	//if err != nil {
	//	logger.Println("Invalid argument: DBMS")
	//	dbChoice = 0
	//}

	// redis
	redisClient, err := boot.InitRedis()
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Println("Connected to redis successfully")

	// tcp server
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Fatalf("cannot create listener: %v", err)
	}

	//jaeger.yaml
	//cfg := &envoy.Configuration{
	//	ServiceName: "publisher",
	//
	//	Sampler: &envoy.SamplerConfig{
	//		Type:  "const",
	//		Param: 0.0001,
	//	},
	//
	//	Reporter: &envoy.ReporterConfig{
	//		LogSpans:           true,
	//		LocalAgentHostPort: "jaeger:6831",
	//	},
	//}

	// tracer
	//tracer, closer, err := cfg.NewTracer(envoy.Logger(jaeger.yaml.StdLogger))
	//if err != nil {
	//	logger.Fatalf("ERROR: cannot init Jaeger: %v\n", err)
	//}
	//defer func(closer io.Closer) {
	//	err := closer.Close()
	//	if err != nil {
	//		logger.Fatalf("ERROR: cannot close tracer: %v\n", err)
	//	}
	//}(closer)
	//opentracing.SetGlobalTracer(tracer)

	// grpc server
	server := grpc.NewServer()
	proto.RegisterBrokerServer(server, &boot.ProtoServer{
		Broker: broker2.NewModule(redisClient, db.POSTGRES),
		Tracer: nil,
	})

	//prometheus
	go boot.InitPrometheus()

	logger.Println("SERVER STARTED SUCCESSFULLY!")

	if err := server.Serve(listener); err != nil {
		logger.Fatalf("Server serve failed: %v", err)
	}
}

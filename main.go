package main

import (
	"context"
	"therealbroker/internal/broker"
	"therealbroker/internal/broker/model"
	"time"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

func main() {
	broker := broker.NewModule()
	ctx := context.Background()

	ch1, _ := broker.Subscribe(ctx, "all")
	id1, _ := broker.Publish(ctx, "all", model.Message{
		0,
		"message1",
		time.Duration(0),
	})
	println((<-ch1).Body)
	msg, err := broker.Fetch(ctx, "all", id1)
	if err != nil {
		println(err.Error())
	}
	println(msg.Body)
	println("___________________________")

	ch2, _ := broker.Subscribe(ctx, "all")
	id2, _ := broker.Publish(ctx, "all", model.Message{
		0,
		"message2",
		time.Duration(500),
	})
	println((<-ch1).Body)
	println((<-ch2).Body)
	msg, err = broker.Fetch(ctx, "all", id2)
	if err != nil {
		println(err.Error())
	}
	println(msg.Body)
	println("___________________________")

	ch3, _ := broker.Subscribe(ctx, "all")
	id3, _ := broker.Publish(ctx, "all", model.Message{
		0,
		"message3",
		time.Duration(500) * time.Second,
	})
	println((<-ch1).Body)
	println((<-ch2).Body)
	println((<-ch3).Body)
	msg, err = broker.Fetch(ctx, "all", id3)
	if err != nil {
		println(err.Error())
	}
	println(msg.Body)
	println("___________________________")

	ch4, _ := broker.Subscribe(ctx, "ali")
	id4, _ := broker.Publish(ctx, "ali", model.Message{
		0,
		"message4",
		time.Duration(500),
	})
	println((<-ch4).Body)
	msg, err = broker.Fetch(ctx, "ali", id4)
	if err != nil {
		println(err.Error())
	}
	println(msg.Body)
	println("___________________________")
}

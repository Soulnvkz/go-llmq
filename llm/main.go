package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wagslane/go-rabbitmq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"sol/proto"
)

func rabbit() {
	conn, err := rabbitmq.NewConn(
		"amqp://admin:admin@localhost",
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	consumer, err := rabbitmq.NewConsumer(
		conn,
		"llm_queue",
		rabbitmq.WithConsumerOptionsRoutingKey("my_routing_key"),
		rabbitmq.WithConsumerOptionsExchangeName("events"),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	err = consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		log.Printf("consumed: %v", string(d.Body))
		// rabbitmq.Ack, rabbitmq.NackDiscard, rabbitmq.NackRequeue
		return rabbitmq.Ack
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// rabbit()
	// Set up a connection to the server.
	conn, err := grpc.NewClient("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewDbeeClient(conn)

	name := "bob"
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &proto.HelloRequest{Name: name})
	if err != nil {
		fmt.Printf("could not greet: %v", err)
	}
	fmt.Printf("Greeting: %s", r.GetMessage())

	// output := func(b []byte, err error) {
	// 	if err != nil {
	// 		fmt.Println("Error: %v", err)
	// 		return
	// 	}
	// 	if b == nil {
	// 		fmt.Println("\nEnd.")
	// 		return
	// 	}

	// 	fmt.Print(string(b))
	// }

	// llm := llm_local.NewLLM()
	// llm.Initilize("/home/sol/programming/ai/models/SAINEMO-reMIX.i1-Q6_K.gguf")
	// defer llm.Clean()

	// err := llm.Proccess("Привет. Как тебя зовут? Расскажи историю про грибы в лесу.", output)
	// if err != nil {
	// 	fmt.Println("Error: %v", err)
	// }

	// err = llm.Proccess("Сколько будет 1 + 1?", output)
	// if err != nil {
	// 	fmt.Println("Error: %v", err)
	// }
}

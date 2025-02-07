package main

import (
	"fmt"
	"sol/llm/llm_local"
)

func main() {
	// Set up a connection to the server.
	// conn, err := grpc.NewClient("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	fmt.Printf("did not connect: %v", err)
	// }
	// defer conn.Close()
	// c := proto.NewDbeeClient(conn)

	// name := "bob"
	// // Contact the server and print out its response.
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// r, err := c.SayHello(ctx, &proto.HelloRequest{Name: name})
	// if err != nil {
	// 	fmt.Printf("could not greet: %v", err)
	// }
	// fmt.Printf("Greeting: %s", r.GetMessage())

	output := func(b []byte, err error) {
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		if b == nil {
			fmt.Print("\nEnd.\n")
			return
		}

		fmt.Print(string(b))
	}

	model := "models/SAINEMO-reMIX.i1-Q6_K.gguf"
	llm := llm_local.NewLLM()
	llm.Initilize(model)
	defer llm.Clean()

	err := llm.Proccess("Привет. Как тебя зовут? Расскажи историю про грибы в лесу.", output)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	err = llm.Proccess("Сколько будет 1 + 1?", output)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

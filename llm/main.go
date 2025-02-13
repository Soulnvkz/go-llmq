package main

import (
	"log"
	"os"
	"sol/llm/llm_local"

	amqp "github.com/rabbitmq/amqp091-go"
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

	model := os.Getenv("MODEL_PATH")
	if len(model) == 0 {
		log.Fatal("ENV MODEL_PATH should be specifed")
	}

	llm := llm_local.NewLLM()
	llm.Initilize(model)
	defer llm.Clean()

	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Panicf("%s: %s", "failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	pubConn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Panicf("%s: %s", "failed to connect to RabbitMQ", err)
	}
	defer pubConn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "failed to create channel", err)
	}
	defer ch.Close()

	cch, err := conn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "failed to create channel", err)
	}
	defer cch.Close()

	pubch, err := pubConn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "failed to create channel", err)
	}
	defer pubch.Close()

	q, err := ch.QueueDeclare(
		"llm_q", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		log.Panicf("%s: %s", "failed to declare requests queue", err)
	}

	qCancel, err := cch.QueueDeclare(
		"llm_q_cancel", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		log.Panicf("%s: %s", "failed to declare cancel queue", err)
	}

	requests, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Panicf("%s: %s", "failed to init consume proccess", err)
	}

	cRequests, err := cch.Consume(
		qCancel.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Panicf("%s: %s", "failed to init consume proccess", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	var forever chan struct{}
	go func() {
		for {
			select {
			case cRequest := <-cRequests:
				log.Printf("Received a cancel request %s", cRequest.Body)
				llm.Cancel()
				cRequest.Ack(false)

			case request := <-requests:
				go func() {
					log.Printf("Received a message: %s", request.Body)
					err = pubch.Publish(
						"",              // exchange
						request.ReplyTo, // routing key
						false,           // mandatory
						false,           // immediate
						amqp.Publishing{
							ContentType:   "text/plain",
							CorrelationId: request.CorrelationId,
							Body:          []byte("<start>"),
						})
					if err != nil {
						log.Printf("Error: %s", err)
					}
					log.Printf("Publish <start>")

					output := func(b []byte, err error) {
						if err != nil {
							log.Printf("Error: %s\n", err)
							return
						}
						if b == nil {
							err = pubch.Publish(
								"",              // exchange
								request.ReplyTo, // routing key
								false,           // mandatory
								false,           // immediate
								amqp.Publishing{
									ContentType:   "text/plain",
									CorrelationId: request.CorrelationId,
									Body:          []byte("<end>"),
								})
							if err != nil {
								log.Printf("Error: %s", err)
							}
							log.Printf("Publish <end>")

							return
						}

						err = pubch.Publish(
							"",              // exchange
							request.ReplyTo, // routing key
							false,           // mandatory
							false,           // immediate
							amqp.Publishing{
								ContentType:   "text/plain",
								CorrelationId: request.CorrelationId,
								Body:          b,
							})
						if err != nil {
							log.Printf("Error: %s", err)
						}
						log.Printf("Publish next")
					}

					err = llm.Proccess(string(request.Body), output)
					if err != nil {
						log.Printf("Error: %s", err)
					}

					request.Ack(false)
				}()
			}
		}

		// for request := range requests {
		// 	log.Printf("Received a message: %s", request.Body)
		// 	err = pubch.Publish(
		// 		"",              // exchange
		// 		request.ReplyTo, // routing key
		// 		false,           // mandatory
		// 		false,           // immediate
		// 		amqp.Publishing{
		// 			ContentType:   "text/plain",
		// 			CorrelationId: request.CorrelationId,
		// 			Body:          []byte("<start>"),
		// 		})
		// 	if err != nil {
		// 		log.Printf("Error: %s", err)
		// 	}

		// 	output := func(b []byte, err error) {
		// 		if err != nil {
		// 			log.Printf("Error: %s\n", err)
		// 			return
		// 		}
		// 		if b == nil {
		// 			err = pubch.Publish(
		// 				"",              // exchange
		// 				request.ReplyTo, // routing key
		// 				false,           // mandatory
		// 				false,           // immediate
		// 				amqp.Publishing{
		// 					ContentType:   "text/plain",
		// 					CorrelationId: request.CorrelationId,
		// 					Body:          []byte("<end>"),
		// 				})
		// 			if err != nil {
		// 				log.Printf("Error: %s", err)
		// 			}

		// 			return
		// 		}

		// 		log.Print(string(b))

		// 		err = pubch.Publish(
		// 			"",              // exchange
		// 			request.ReplyTo, // routing key
		// 			false,           // mandatory
		// 			false,           // immediate
		// 			amqp.Publishing{
		// 				ContentType:   "text/plain",
		// 				CorrelationId: request.CorrelationId,
		// 				Body:          b,
		// 			})
		// 		if err != nil {
		// 			log.Printf("Error: %s", err)
		// 		}
		// 	}

		// 	err = llm.Proccess(string(request.Body), output)
		// 	if err != nil {
		// 		log.Printf("Error: %s", err)
		// 	}

		// 	request.Ack(false)
		// }
	}()
	<-forever

	// err = llm.Proccess("Привет. Как тебя зовут? Расскажи историю про грибы в лесу.", output)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// }

	// err = llm.Proccess("Сколько будет 1 + 1?", output)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// }
}

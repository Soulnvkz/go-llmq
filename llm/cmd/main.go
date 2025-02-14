package main

import (
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soulnvkz/llm/internal/llama"
	"github.com/soulnvkz/mq"
)

func Getenv(env string) (v string) {
	v, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("ENV %s should be specifed", env)
	}
	return
}

func main() {
	model := Getenv("MODEL_PATH")

	mq_user := Getenv("MQ_USER")
	mq_password := Getenv("MQ_PASSWORD")
	mq_host := Getenv("MQ_HOST")
	mq_port := Getenv("MQ_PORT")

	llm := llama.NewLLM()
	llm.Initilize(model)
	defer llm.Clean()

	qconn, err := mq.MQConnect(mq_user, mq_password, mq_host, mq_port, 20)
	if err != nil {
		log.Panicf("%s, failed to connect to RabbitMQ", err)
	}
	defer qconn.Close()

	pqconn, err := mq.MQConnect(mq_user, mq_password, mq_host, mq_port, 20)
	if err != nil {
		log.Panicf("%s, failed to connect to RabbitMQ", err)
	}
	defer pqconn.Close()

	req_channel, err := qconn.Channel()
	if err != nil {
		log.Panicf("%s, failed declare channel", err)
	}
	defer req_channel.Close()
	err = req_channel.Qos(1, 0, false)
	if err != nil {
		log.Panicf("%s, failed to setup qos", err)
	}

	cancel_channel, err := qconn.Channel()
	if err != nil {
		log.Panicf("%s, failed declare channel", err)
	}
	defer cancel_channel.Close()

	err = cancel_channel.ExchangeDeclare(
		"llm_cancel_ex",
		"fanout",
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Panicf("%s, failed declare channel", err)
	}

	pub_channel, err := pqconn.Channel()
	if err != nil {
		log.Panicf("%s, failed declare channel", err)
	}
	defer pub_channel.Close()

	reqq, err := mq.NewMQueue(req_channel, "llm_q")
	if err != nil {
		log.Panicf("%s, failed to declare queue", err)
	}

	cancelq, err := mq.NewMQueue(cancel_channel, "llm_cancel")
	if err != nil {
		log.Panicf("%s, failed to declare queue", err)
	}
	err = cancel_channel.QueueBind(cancelq.Name(), "", "llm_cancel_ex", false, nil)
	if err != nil {
		log.Panicf("%s, failed to bind queue", err)
	}

	llm_r, err := reqq.Consume()
	if err != nil {
		log.Panicf("%s, failed to start consume", err)
	}

	llm_cancel, err := cancelq.Consume()
	if err != nil {
		log.Panicf("%s, failed to start consume", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			cRequest := <-llm_cancel
			log.Printf("Received a cancel request %s", cRequest.CorrelationId)
			llm.Cancel(cRequest.CorrelationId)
			cRequest.Ack(false)
		}
	}()

	go func() {
		for {
			request := <-llm_r
			log.Printf("Received a message: %s", request.Body)
			err = pub_channel.Publish(
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
				request.Ack(false)
				continue
			}
			log.Printf("Publish <start>")

			output := func(b []byte, err error) error {
				if err != nil {
					log.Printf("Error: %s\n", err)
					return nil
				}

				err = pub_channel.Publish(
					"",              // exchange
					request.ReplyTo, // routing key
					false,           // mandatory
					false,           // immediate
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: request.CorrelationId,
						Body:          b,
					})
				log.Printf("next")
				if err != nil {
					log.Printf("Error: %s", err)
					return err
				}

				return nil
			}

			stop, err := llm.ProccessNext(string(request.Body), request.CorrelationId, output)
			if err != nil {
				log.Printf("Error: %s", err)
				request.Ack(false)
				continue
			}
			<-stop

			err = pub_channel.Publish(
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
				continue
			}
			log.Printf("Publish <end>")

			request.Ack(false)
		}
	}()

	wg.Wait()
}

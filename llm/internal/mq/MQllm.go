package mq

import (
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soulnvkz/llm/internal/llama"
	"github.com/soulnvkz/mq"
	"github.com/soulnvkz/mq/domain"
)

type MQConfig struct {
	CancelExKey string
	ReqQKey     string
}

type MQllm struct {
	reqChannel    *amqp.Channel
	cancelChannel *amqp.Channel
	pubChannel    *amqp.Channel

	reqQ    *mq.MQQueue
	cancelQ *mq.MQQueue

	config MQConfig
}

func NewMQllm(pull, pub *mq.MQConnection, config MQConfig) (*MQllm, error) {
	// initilize channels
	req_channel, err := pull.Channel()
	if err != nil {
		return nil, err
	}
	err = req_channel.Qos(1, 0, false)
	if err != nil {
		return nil, err
	}
	cancel_channel, err := pull.Channel()
	if err != nil {
		return nil, err

	}
	pub_channel, err := pub.Channel()
	if err != nil {
		return nil, err
	}

	err = cancel_channel.ExchangeDeclare(
		config.CancelExKey,
		"fanout",
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	cancelQ, err := mq.NewMQueue(cancel_channel, "")
	if err != nil {
		return nil, err
	}
	err = cancel_channel.QueueBind(cancelQ.Name(), "", config.CancelExKey, false, nil)
	if err != nil {
		return nil, err
	}

	reqQ, err := mq.NewMQueue(req_channel, config.ReqQKey)
	if err != nil {
		return nil, err
	}

	return &MQllm{
		reqChannel:    req_channel,
		cancelChannel: cancel_channel,
		pubChannel:    pub_channel,

		reqQ:    reqQ,
		cancelQ: cancelQ,

		config: config,
	}, nil
}

func (llmq *MQllm) reply(replyTo string, resp domain.CompletionsResponse) error {
	buff, err := resp.Marshal()
	if err != nil {
		return err
	}

	err = llmq.pubChannel.Publish(
		"",      // exchange
		replyTo, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: resp.RequestID,
			Body:          buff,
		})
	if err != nil {
		return err
	}

	return nil
}

func (llmq *MQllm) Consume(llm *llama.LLM) error {
	llm_r, err := llmq.reqQ.Consume()
	if err != nil {
		return err
	}

	llm_cancel, err := llmq.cancelQ.Consume()
	if err != nil {
		return err
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
			err = llmq.reply(request.ReplyTo, domain.CompletionsResponse{
				RequestID: request.CorrelationId,
				ResType:   domain.CompletionsStart,
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

				err = llmq.reply(request.ReplyTo, domain.CompletionsResponse{
					RequestID: request.CorrelationId,
					Content:   string(b),
					ResType:   domain.CompletionsNext,
				})
				log.Printf("next")
				if err != nil {
					log.Printf("Error: %s", err)
					return err
				}

				return nil
			}
			req := domain.CompletionsRequest{}
			req.UnMarshal(request.Body)
			stop, err := llm.ProccessNext(req.Content, request.CorrelationId, output)
			if err != nil {
				log.Printf("Error: %s", err)
				request.Ack(false)
				continue
			}
			<-stop
			err = llmq.reply(request.ReplyTo, domain.CompletionsResponse{
				RequestID: request.CorrelationId,
				ResType:   domain.CompletionsEnd,
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
	return nil
}

func (llmq *MQllm) Close() {
	llmq.reqChannel.Close()
	llmq.cancelChannel.Close()
	llmq.pubChannel.Close()
}

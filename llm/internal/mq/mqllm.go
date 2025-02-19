package mq

import (
	"context"
	"log"

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

func (llmq *MQllm) ConsumeCompletionsRequests(ctx context.Context, pbuilder llama.PromptBuilder, d ResponseGenerator) (<-chan bool, error) {
	llm_r, err := llmq.reqQ.Consume()
	if err != nil {
		return nil, err
	}
	done := make(chan bool)

	go func() {
	main_loop:
		for {
			select {
			case <-ctx.Done():
				done <- true
				break main_loop
			case req := <-llm_r:
				req.Ack(false)

				cr := domain.CompletionsRequest{}
				err := cr.UnMarshal(req.Body)
				if err != nil {
					log.Printf("%s, unsupported request data", err)
					continue
				}

				err = llmq.reply(req.ReplyTo, domain.CompletionsResponse{
					RequestID: req.CorrelationId,
					ResType:   domain.CompletionsStart,
				})
				if err != nil {
					log.Printf("%s, failed to reply", err)
					continue
				}

				prompt, err := pbuilder.Build(cr.ChatMessages, cr.Content)
				if err != nil {
					log.Printf("%s, failed to build prompt", err)
					continue
				}

				req_ctx, cancel := context.WithCancel(ctx)
				next, stop, err := d.Proccess(req_ctx, prompt, cr.RequestID)
				if err != nil {
					log.Printf("%s, failed to start generation", err)
					cancel()
					continue
				}
			proccess_loop:
				for {
					select {
					case <-stop:
						log.Printf("%s stop", req.CorrelationId)
						err = llmq.reply(req.ReplyTo, domain.CompletionsResponse{
							RequestID: req.CorrelationId,
							ResType:   domain.CompletionsEnd,
						})
						if err != nil {
							log.Printf("%s, failed to reply", err)
							cancel()
							break proccess_loop
						}
						break proccess_loop
					case buff := <-next:
						err = llmq.reply(req.ReplyTo, domain.CompletionsResponse{
							RequestID: req.CorrelationId,
							Content:   string(buff),
							ResType:   domain.CompletionsNext,
						})
						if err != nil {
							log.Printf("%s, failed to reply", err)
							cancel()
							break proccess_loop
						}
					}
				}
				cancel()
			}
		}
	}()

	return done, nil
}

func (llmq *MQllm) ConsumeCancellations(ctx context.Context, c ResponseCancellation) (<-chan bool, error) {
	llm_cancel, err := llmq.cancelQ.Consume()
	if err != nil {
		return nil, err
	}
	done := make(chan bool)
	go func() {
	main_loop:
		for {
			select {
			case <-ctx.Done():
				done <- true
				break main_loop
			case req := <-llm_cancel:
				log.Printf("%s queue cancellation", req.CorrelationId)
				c.Cancel(req.CorrelationId)
				req.Ack(false)
			}
		}
	}()

	return done, nil
}

func (llmq *MQllm) Close() {
	llmq.reqChannel.Close()
	llmq.cancelChannel.Close()
	llmq.pubChannel.Close()
}

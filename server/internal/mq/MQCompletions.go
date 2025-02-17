package mq

import (
	"context"
	"errors"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soulnvkz/log"
	"github.com/soulnvkz/mq"
	domain "github.com/soulnvkz/mq/domain"
)

const (
	CancelExchangeKey = "llm_cancel_ex"
	PubQueueKey       = "llm_q"
)

type Consumer interface {
	OnDone() error
	OnNext(r domain.CompletionsResponse) error
}

type MQCompletions struct {
	completionsChannel *amqp.Channel
	publishChannel     *amqp.Channel
}

func NewMQCompletions(pull, pub *mq.MQConnection) (*MQCompletions, error) {
	completionsChannel, err := pull.Channel()
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to create completions channel"))
	}

	publishChannel, err := pub.Channel()
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to create publish channel"))
	}

	err = publishChannel.ExchangeDeclare(
		CancelExchangeKey,
		"fanout",
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to declare cancel request exchange"))
	}

	return &MQCompletions{
		completionsChannel: completionsChannel,
		publishChannel:     publishChannel,
	}, nil
}

func (comp *MQCompletions) Close() {
	comp.completionsChannel.Close()
	comp.publishChannel.Close()
}

func (comp *MQCompletions) NewCompletionsQueue() (*mq.MQQueue, error) {
	q, err := mq.NewMQueue(comp.completionsChannel, "")
	if err != nil {
		return nil, err
	}

	return q, nil
}

func (comp *MQCompletions) ConsumeCompletions(ctx context.Context, q *mq.MQQueue, c Consumer) error {
	deliveries, err := q.Consume()
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				if err := c.OnDone(); err != nil {
					log.Error().Print(err)
				}
				break loop
			case next := <-deliveries:
				resp := &domain.CompletionsResponse{
					RequestID: next.CorrelationId,
				}
				err := resp.UnMarshal(next.Body)
				if err != nil {
					log.Error().Printf("unsupported mq message")
					continue loop
				}
				if err = c.OnNext(*resp); err != nil {
					break loop
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (comp *MQCompletions) RequestCompletions(ctx context.Context, q *mq.MQQueue, req domain.CompletionsRequest) error {
	buff, err := req.Marshal()
	if err != nil {
		return err
	}

	err = comp.publishChannel.PublishWithContext(ctx,
		"",          // exchange
		PubQueueKey, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: req.RequestID,
			ReplyTo:       q.Name(),
			Body:          []byte(buff),
		})
	if err != nil {
		return err
	}

	return nil
}

func (comp *MQCompletions) CancelRequest(requestID string) error {
	err := comp.publishChannel.Publish(
		CancelExchangeKey, // exchange
		"",                // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: requestID,
			Body:          []byte{},
		})
	if err != nil {
		return err
	}

	return nil
}

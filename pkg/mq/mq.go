package mq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQConnection struct {
	*amqp.Connection
}

func MQConnect(user, passw, host, port string, limit int) (*MQConnection, error) {
	connstr := fmt.Sprintf("amqp://%s:%s@%s:%s", user, passw, host, port)

	retry := 0
	for {
		conn, err := amqp.Dial(connstr)
		if err != nil {
			log.Printf("%s, failed to dial amqp with %d retry", err, retry)
			retry++
			if retry > limit {
				return nil, err
			}

			time.Sleep(500 * time.Millisecond)

			continue
		}
		return &MQConnection{
			conn,
		}, nil
	}
}

type MQQueue struct {
	ch *amqp.Channel
	q  *amqp.Queue
}

func NewMQueue(ch *amqp.Channel, name string) (*MQQueue, error) {
	queue, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	return &MQQueue{
		ch: ch,
		q:  &queue,
	}, nil
}

func (q *MQQueue) Name() string {
	return q.q.Name
}

func (q *MQQueue) Consume() (req <-chan amqp.Delivery, err error) {
	req, err = q.ch.Consume(
		q.q.Name, // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	return
}

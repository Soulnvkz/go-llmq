package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soulnvkz/log"
	"github.com/soulnvkz/mq"
)

type Message struct {
	MessageType int    `json:"message_type"`
	Content     string `json:"content,omitempty"`
}

const (
	PingMessage = 1
	PongMessage = 1

	CompletitionsMessage = 2
	CancelMessage        = 3

	CompletitionsStart = 2
	CompletitionsNext  = 3
	CompletitionsEnd   = 4
	CompletitionsQueue = 5

	Error = 6
)

type WSCompletions struct {
	c      *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	pingTicker *time.Ticker

	mu           *sync.Mutex
	streamCancel *context.CancelFunc

	completions_channel *amqp.Channel
	publish_channel     *amqp.Channel
}

const (
	PING_DELAY = 5000 * time.Minute
)

func NewWSCompletions(
	c *websocket.Conn,
	completitons_channel, publish_channel *amqp.Channel,
	ctx context.Context) *WSCompletions {
	nctx, cancel := context.WithCancel(ctx)

	return &WSCompletions{
		c:                   c,
		ctx:                 nctx,
		cancel:              cancel,
		pingTicker:          time.NewTicker(PING_DELAY),
		completions_channel: completitons_channel,
		publish_channel:     publish_channel,
		mu:                  &sync.Mutex{},
		streamCancel:        nil,
	}
}

func (socket *WSCompletions) Close() {
	log.Info().Printf("closing...")
	socket.c.Close()
}

func (socket *WSCompletions) HandleMessages() {
loop:
	for {
		out := socket.readNext()
		select {
		case <-socket.ctx.Done():
			break loop
		case message := <-out:
			go socket.handleMessage(message)
		case <-socket.pingTicker.C:
			log.Info().Printf("no pings...closing connections")
			socket.cancel()
		}
	}
}

func (socket *WSCompletions) handlePing() {
	socket.pingTicker.Reset(PING_DELAY)
	err := socket.writePong()
	if err != nil {
		log.Error().Printf("writing message error: %s", err)
		socket.cancel()
	}
}

func (socket *WSCompletions) handleStreamCancel() {
	socket.mu.Lock()
	if socket.streamCancel != nil {
		(*socket.streamCancel)()
		socket.streamCancel = nil
	} else {
		socket.writeError(errors.New("no active completions"))
	}
	socket.mu.Unlock()
}

func (socket *WSCompletions) handleCompletions(message *Message) {
	if len(message.Content) == 0 {
		log.Info().Println("content is required for completitions")
		err := socket.writeError(errors.New("content is required for completitions"))
		if err != nil {
			socket.cancel()
		}
		return
	}

	socket.mu.Lock()
	if socket.streamCancel != nil {
		socket.mu.Unlock()
		log.Info().Println("previous stream is not finished")
		err := socket.writeError(errors.New("previous stream is not finished"))
		if err != nil {
			socket.cancel()
		}
		return
	}
	socket.mu.Unlock()

	go func() {
		socket.mu.Lock()
		ctx, cancel := context.WithCancel(socket.ctx)
		socket.streamCancel = &cancel
		socket.mu.Unlock()

		defer func() {
			socket.mu.Lock()
			if socket.streamCancel != nil {
				socket.streamCancel = nil
			}
			socket.mu.Unlock()
		}()

		respq, err := mq.NewMQueue(socket.completions_channel, "")
		if err != nil {
			log.Error().Printf("failed to declare queue")
			return
		}

		responses, err := respq.Consume()
		if err != nil {
			log.Error().Printf("failed initilize consume %s", err)
			return
		}

		err = socket.writeQueueCompletions()
		if err != nil {
			log.Error().Printf("failed to response %s", err)
			return
		}

		request_id := uuid.New().String()

		err = socket.publish_channel.ExchangeDeclare(
			"llm_cancel_ex",
			"fanout",
			false,
			true,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Error().Printf("failed declare channel %s", err)
			return
		}

		err = socket.publish_channel.PublishWithContext(ctx,
			"",      // exchange
			"llm_q", // routing key
			false,   // mandatory
			false,   // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: request_id,
				ReplyTo:       respq.Name(),
				Body:          []byte(message.Content),
			})
		if err != nil {
			log.Error().Printf("failed publish %s", err)
			return
		}

	loop:
		for {
			select {
			case <-ctx.Done():
				// proccess already started and can be canceled exacly that proccess
				err = socket.publish_channel.PublishWithContext(ctx,
					"llm_cancel_ex", // exchange
					"",              // routing key
					false,           // mandatory
					false,           // immediate
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: request_id,
						Body:          []byte(message.Content),
					})
				if err != nil {
					log.Error().Printf("failed publish %s", err)
					return
				}
				break loop
			case d := <-responses:
				if request_id == d.CorrelationId {
					str := string(d.Body)
					if str == "<start>" {
						err = socket.writeStartCompletions()
						if err != nil {
							break loop
						}
						continue loop
					}
					if str == "<end>" {
						socket.writeEndCompletions()
						break loop
					}

					err = socket.writeCompletions(d.Body)
					if err != nil {
						break loop
					}
				}
			}
		}
	}()
}

func (socket *WSCompletions) handleMessage(buff []byte) {
	var message Message
	err := json.Unmarshal(buff, &message)
	if err != nil {
		log.Info().Printf("unssuported message")
		err = socket.writeError(errors.New("unssuported message"))
		if err != nil {
			socket.cancel()
		}
		return
	}

	switch {
	case message.MessageType == PingMessage:
		socket.handlePing()
	case message.MessageType == CancelMessage:
		socket.handleStreamCancel()
	case message.MessageType == CompletitionsMessage:
		socket.handleCompletions(&message)
	default:
		log.Info().Printf("unssuported message")
		err = socket.writeError(errors.New("unssuported message"))
		if err != nil {
			socket.cancel()
		}
		return
	}
}

func (socket *WSCompletions) writePong() error {
	message := &Message{
		MessageType: PongMessage,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *WSCompletions) writeCompletions(buff []byte) error {
	message := &Message{
		MessageType: CompletitionsNext,
		Content:     string(buff),
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *WSCompletions) writeQueueCompletions() error {
	message := &Message{
		MessageType: CompletitionsQueue,
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *WSCompletions) writeStartCompletions() error {
	message := &Message{
		MessageType: CompletitionsStart,
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *WSCompletions) writeEndCompletions() error {
	message := &Message{
		MessageType: CompletitionsEnd,
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *WSCompletions) writeError(err error) error {
	message := &Message{
		MessageType: Error,
		Content:     err.Error(),
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *WSCompletions) readNext() chan []byte {
	buff := make(chan []byte)

	go func() {
		mt, message, err := socket.c.ReadMessage()
		if err != nil {
			log.Error().Println("reading message error:", err)
			socket.cancel()
			return
		}
		if mt == websocket.CloseMessage {
			log.Error().Print("requesting close connection...")
			socket.cancel()
			return
		}

		buff <- message
	}()

	return buff
}

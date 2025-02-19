package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/soulnvkz/log"
	domain "github.com/soulnvkz/mq/domain"
	mqc "github.com/soulnvkz/server/internal/mq"
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

	mqcompletions *mqc.MQCompletions
}

const (
	PING_DELAY = 120 * time.Second
)

func NewWSCompletions(
	ctx context.Context,
	c *websocket.Conn,
	mqcomp *mqc.MQCompletions) *WSCompletions {
	nctx, cancel := context.WithCancel(ctx)

	return &WSCompletions{
		c:             c,
		ctx:           nctx,
		cancel:        cancel,
		pingTicker:    time.NewTicker(PING_DELAY),
		mqcompletions: mqcomp,
		mu:            &sync.Mutex{},
		streamCancel:  nil,
	}
}

func (socket *WSCompletions) Close() {
	log.Info().Printf("closing...")
	socket.c.Close()
}

func (socket *WSCompletions) HandleMessages() {
loop:
	for {
		buff := socket.readNext()
		if buff != nil {
			go socket.handleMessage(buff)
		}

		select {
		case <-socket.ctx.Done():
			break loop
		case <-socket.pingTicker.C:
			log.Info().Printf("no pings...closing connections")
			socket.cancel()
		default:
			continue loop
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

		q, err := socket.mqcompletions.NewCompletionsQueue()
		if err != nil {
			log.Error().Printf("failed to declare queue, %s", err)
			return
		}

		err = socket.writeQueueCompletions()
		if err != nil {
			log.Error().Printf("failed to response, %s", err)
			return
		}

		request_id := uuid.New().String()
		request := domain.CompletionsRequest{
			RequestID: request_id,
			Content:   message.Content,
		}

		err = socket.mqcompletions.RequestCompletions(ctx, q, request)
		if err != nil {
			log.Error().Printf("failed publish, %s", err)
			return
		}

		err = socket.mqcompletions.ConsumeCompletions(ctx, q, &WSConsumer{
			requestID: request_id,
			socket:    socket,
		})
		if err != nil {
			log.Error().Printf("failed to start consume, %s", err)
			return
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

	return socket.writeNext(data)
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

	return socket.writeNext(data)
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

	return socket.writeNext(data)
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
	return socket.writeNext(data)
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

	return socket.writeNext(data)
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

	return socket.writeNext(data)
}

func (socket *WSCompletions) readNext() []byte {
	mt, buff, err := socket.c.ReadMessage()

	if err != nil {
		log.Error().Println("reading message error:", err)
		socket.cancel()
		return nil
	}
	if mt == websocket.CloseMessage {
		log.Info().Print("requested close connection...")
		socket.cancel()
		return nil
	}

	return buff
}

func (socket *WSCompletions) writeNext(data []byte) error {
	socket.mu.Lock()
	defer socket.mu.Unlock()
	return socket.c.WriteMessage(websocket.TextMessage, data)
}

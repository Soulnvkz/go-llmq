package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	MessageType int    `json:"message_type"`
	Content     string `json:"content,omitempty"`
}

const (
	PingMessage          = 1
	CompletitionsMessage = 2
	CancelMessage        = 3
	CompletitionsStart   = 1
	CompletitionsNext    = 2
	CompletitionsEnd     = 3
	CompletitionsQueue   = 4
	Error                = 5
	Lorem                = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
)

type Socket struct {
	c      *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc

	pingTicker *time.Ticker

	mu           *sync.Mutex
	streamCancel *context.CancelFunc
}

const (
	PING_DELAY = 5000 * time.Minute
)

func NewSocket(c *websocket.Conn, ctx context.Context) *Socket {
	nctx, cancel := context.WithCancel(ctx)

	return &Socket{
		c:            c,
		ctx:          nctx,
		cancel:       cancel,
		pingTicker:   time.NewTicker(PING_DELAY),
		mu:           &sync.Mutex{},
		streamCancel: nil,
	}
}

func (socket Socket) Close() {
	log.Printf("closing...")
	socket.c.Close()
}

func (socket *Socket) HandleMessages() {
loop:
	for {
		out := socket.readNext()
		select {
		case <-socket.ctx.Done():
			break loop
		case message := <-out:
			go socket.handleMessage(message)
		case <-socket.pingTicker.C:
			log.Printf("no pings...closing connections")
			socket.cancel()
		}
	}
}

func (socket *Socket) handlePing() {
	socket.pingTicker.Reset(PING_DELAY)
	err := socket.c.WriteMessage(websocket.TextMessage, []byte("pong"))
	if err != nil {
		log.Println("writing message error:", err)
		socket.cancel()
	}
}

func (socket *Socket) handleStreamCancel() {
	socket.mu.Lock()
	if socket.streamCancel != nil {
		(*socket.streamCancel)()
		socket.streamCancel = nil
	} else {
		socket.writeError(errors.New("no active completions"))
	}
	socket.mu.Unlock()
}

func (socket *Socket) handleCompletions(message *Message) {
	if len(message.Content) == 0 {
		log.Println("content is required for completitions")
		err := socket.writeError(errors.New("content is required for completitions"))
		if err != nil {
			socket.cancel()
		}
		return
	}

	socket.mu.Lock()
	if socket.streamCancel != nil {
		socket.mu.Unlock()
		log.Println("previous stream is not finished")
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

		stream := strings.NewReader(Lorem)
		done := make(chan bool)
		read := make(chan bool)

	loop:
		for {
			go func() {
				abuff := make([]byte, 8)
				_, err := stream.Read(abuff)
				if err == io.EOF {
					done <- true
					return
				}
				if len(abuff) > 0 {
					err = socket.writeCompletions(abuff)
					if err != nil {
						done <- true
					}
				}
				time.Sleep(1000 * time.Millisecond)
				read <- true
			}()
			select {
			case <-ctx.Done():
				break loop
			case <-read:
			case <-done:
				break loop
			}
		}
	}()
}

func (socket *Socket) handleMessage(buff []byte) {
	var message Message
	err := json.Unmarshal(buff, &message)
	if err != nil {
		log.Printf("unssuported message")
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
		log.Printf("unssuported message")
		err = socket.writeError(errors.New("unssuported message"))
		if err != nil {
			socket.cancel()
		}
		return
	}
}

func (socket *Socket) writeCompletions(buff []byte) error {
	message := &Message{
		MessageType: CompletitionsNext,
		Content:     string(buff),
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *Socket) writeError(err error) error {
	message := &Message{
		MessageType: Error,
		Content:     err.Error(),
	}
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("failed to marshal message, %s", err)
		return err
	}

	return socket.c.WriteMessage(websocket.TextMessage, data)
}

func (socket *Socket) readNext() chan []byte {
	buff := make(chan []byte)

	go func() {
		mt, message, err := socket.c.ReadMessage()
		if err != nil {
			log.Println("reading message error:", err)
			socket.cancel()
			return
		}
		if mt == websocket.CloseMessage {
			log.Print("requesting close connection...")
			socket.cancel()
			return
		}

		buff <- message
	}()

	return buff
}

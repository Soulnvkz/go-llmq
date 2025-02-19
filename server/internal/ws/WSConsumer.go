package ws

import (
	"io"

	"github.com/soulnvkz/log"
	"github.com/soulnvkz/mq/domain"
)

type WSConsumer struct {
	requestID string
	socket    *WSCompletions
	message   []byte
	userm     []byte
}

func NewWSConsumer(reqID string, s *WSCompletions, userm []byte) *WSConsumer {
	return &WSConsumer{
		requestID: reqID,
		socket:    s,
		userm:     userm,
		message:   make([]byte, 0, 1024),
	}
}

func (c *WSConsumer) OnDone() error {
	log.Info().Printf("call OnDone, %s", c.requestID)
	if err := c.socket.mqcompletions.CancelRequest(c.requestID); err != nil {
		return err
	}
	return nil
}

func (c *WSConsumer) OnNext(r domain.CompletionsResponse) error {
	log.Info().Printf("call OnNext %s", r.RequestID)
	switch {
	case r.ResType == domain.CompletionsStart:
		if err := c.socket.writeStartCompletions(); err != nil {
			return err
		}
	case r.ResType == domain.CompletionsEnd:
		c.socket.chat.Add(domain.ChatMessage{
			Role:    "user",
			Content: string(c.userm),
		})
		c.socket.chat.Add(domain.ChatMessage{
			Role:    "assistant",
			Content: string(c.message),
		})
		c.socket.writeEndCompletions()
		return io.EOF
	case r.ResType == domain.CompletionsNext:
		buff := []byte(r.Content)
		c.message = append(c.message, buff...)
		if err := c.socket.writeCompletions(buff); err != nil {
			return err
		}
	default:
		log.Info().Printf("unsuported completions response type, %v", r)
	}
	return nil
}

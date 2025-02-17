package ws

import (
	"io"

	"github.com/soulnvkz/log"
	"github.com/soulnvkz/mq/domain"
)

type WSConsumer struct {
	requestID string
	socket    *WSCompletions
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
		c.socket.writeEndCompletions()
		return io.EOF
	case r.ResType == domain.CompletionsNext:
		if err := c.socket.writeCompletions([]byte(r.Content)); err != nil {
			return err
		}
	default:
		log.Info().Printf("unsuported completions response type, %v", r)
	}
	return nil
}

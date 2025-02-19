package chat

import "github.com/soulnvkz/mq/domain"

type ChatContext struct {
	Messages []domain.ChatMessage
}

func NewChatContext() *ChatContext {
	return &ChatContext{
		Messages: make([]domain.ChatMessage, 0, 50),
	}
}

func (c *ChatContext) Add(m domain.ChatMessage) {
	c.Messages = append(c.Messages, m)
}

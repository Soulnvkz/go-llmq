package llama

import (
	"errors"
	"fmt"

	"github.com/soulnvkz/mq/domain"
)

type PromptBuilder interface {
	Build(messages []domain.ChatMessage, next string) (string, error)
}

type LLMPromptBuilder struct {
	llm *LLM
}

func NewPromptBuilder(llm *LLM) LLMPromptBuilder {
	return LLMPromptBuilder{
		llm: llm,
	}
}

func (b LLMPromptBuilder) promptFromModelChatTemplate(messages []domain.ChatMessage) (string, error) {
	i := 0
	for {
		p, err := b.llm.ApplyChatTemplate(messages)
		if err != nil {
			return "", err
		}
		l, _, err := b.llm.tokenizePrompt(p)
		if err != nil {
			return "", err
		}
		if l <= (b.llm.n_ctx - b.llm.n_predict) {
			return p, nil
		}

		i++
		if len(messages) == 0 {
			return "", errors.New("")
		}

		messages = messages[1:]
	}
}

func (b LLMPromptBuilder) promptFromDefaultTemplate(messages []domain.ChatMessage) (string, error) {
	assistant := "<|start_header_id|>assistant<|end_header_id|>"
	buff := make([]byte, len(assistant))
	copy(buff, []byte(assistant))

	for i := len(messages) - 1; i >= 0; i-- {
		var m string
		switch {
		case messages[i].Role == "user":
			m = fmt.Sprintf("<|start_header_id|>user<|end_header_id|>%s<|eot_id|>\n", messages[i].Content)
		case messages[i].Role == "assistant":
			m = fmt.Sprintf("<|start_header_id|>assistant<|end_header_id|>%s<|eot_id|>\n", messages[i].Content)
		default:
			continue
		}

		newbuff := make([]byte, len(m)+len(buff))
		copy(newbuff, []byte(m))
		copy(newbuff[len(m):], buff)
		l, _, err := b.llm.tokenizePrompt(string(newbuff))
		if err != nil {
			return "", err
		}

		if l >= (b.llm.n_ctx - b.llm.n_predict) {
			return string(buff), nil
		}
		buff = newbuff
	}

	return string(buff), nil
}

func (b LLMPromptBuilder) Build(messages []domain.ChatMessage, next string) (string, error) {
	messages = append(messages, domain.ChatMessage{
		Role:    "user",
		Content: next,
	})
	if len(b.llm.model_chat_template) > 0 {
		return b.promptFromModelChatTemplate(messages)
	} else {
		return b.promptFromDefaultTemplate(messages)
	}
}

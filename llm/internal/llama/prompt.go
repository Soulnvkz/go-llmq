package llama

import (
	"fmt"
	"log"

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

func (p LLMPromptBuilder) Build(messages []domain.ChatMessage, next string) (string, error) {
	nextm := fmt.Sprintf("<|start_header_id|>user<|end_header_id|>%s<|eot_id|>\n", next)
	buff := make([]byte, len(nextm))
	copy(buff, []byte(nextm))

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

		newbuff := make([]byte, len(buff)+len(m))
		copy(newbuff, []byte(m))
		copy(newbuff[len(m):], buff)
		l, _, err := p.llm.tokenizePrompt(string(newbuff))
		if err != nil {
			return "", err
		}
		log.Printf("l %d", l)

		if l >= (p.llm.n_ctx - p.llm.n_predict) {
			log.Printf("l %d > max %d, index %d, total %d", l, p.llm.n_ctx-p.llm.n_predict, i, len(messages))
			return string(buff), nil
		}
		buff = newbuff
	}

	return string(buff), nil
}

package domain

import "encoding/json"

const (
	CompletionsStart = 1
	CompletionsNext  = 2
	CompletionsEnd   = 3
)

type CompletionsResponse struct {
	RequestID string `json:"request_id"`
	Content   string `json:"content,omitempty"`
	ChatID    string `json:"chat_id,omitempty"`

	ResType uint8 `json:"response_type"`
}

func (r CompletionsResponse) Marshal() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (r *CompletionsResponse) UnMarshal(data []byte) error {
	err := json.Unmarshal(data, r)
	if err != nil {
		return err
	}
	return nil
}

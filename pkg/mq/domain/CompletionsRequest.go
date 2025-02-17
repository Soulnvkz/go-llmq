package domain

import "encoding/json"

type CompletionsRequest struct {
	RequestID string `json:"request_id"`
	Content   string `json:"content,omitempty"`
	ChatID    string `json:"chat_id,omitempty"`
}

func (r CompletionsRequest) Marshal() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (r *CompletionsRequest) UnMarshal(data []byte) error {
	err := json.Unmarshal(data, r)
	if err != nil {
		return err
	}
	return nil
}

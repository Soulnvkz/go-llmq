package mq

import "context"

type ResponseGenerator interface {
	Proccess(ctx context.Context, prompt string, req string) (chan []byte, chan bool, error)
}

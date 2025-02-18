package mq

type ResponseCancellation interface {
	Cancel(req string)
}

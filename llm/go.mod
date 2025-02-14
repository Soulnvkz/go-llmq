module github.com/soulnvkz/llm

go 1.23.5

replace github.com/soulnvkz/mq => ../pkg/mq

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/soulnvkz/mq v0.0.0-00010101000000-000000000000
)

module github.com/soulnvkz/server

go 1.23.5

replace github.com/soulnvkz/log => ../pkg/log

replace github.com/soulnvkz/mq => ../pkg/mq

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/soulnvkz/log v0.0.0-00010101000000-000000000000
	github.com/soulnvkz/mq v0.0.0-00010101000000-000000000000
)

module sol/llm

go 1.23.5

replace sol/proto => ../proto

require (
	github.com/wagslane/go-rabbitmq v0.15.0
	google.golang.org/grpc v1.70.0
	sol/proto v0.0.0-00010101000000-000000000000
)

require (
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)

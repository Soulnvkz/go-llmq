package main

import (
	"context"
	"log"
	"os"

	"github.com/soulnvkz/llm/internal/llama"
	mqc "github.com/soulnvkz/llm/internal/mq"
	"github.com/soulnvkz/mq"
)

func Getenv(env string) (v string) {
	v, f := os.LookupEnv(env)
	if !f {
		log.Fatalf("ENV %s should be specifed", env)
	}
	return
}

func main() {
	model := Getenv("MODEL_PATH")

	mq_user := Getenv("MQ_USER")
	mq_password := Getenv("MQ_PASSWORD")
	mq_host := Getenv("MQ_HOST")
	mq_port := Getenv("MQ_PORT")

	mq_cancel_ex := Getenv("MQ_CANCEL_EX")
	mq_llm_q := Getenv("MQ_LLM_Q")

	llm := llama.NewLLM(context.Background())
	llm.Initilize(model)
	defer llm.Clean()

	qconn, err := mq.MQConnect(mq_user, mq_password, mq_host, mq_port, 20)
	if err != nil {
		log.Panicf("%s, failed to connect to RabbitMQ", err)
	}
	defer qconn.Close()

	pqconn, err := mq.MQConnect(mq_user, mq_password, mq_host, mq_port, 20)
	if err != nil {
		log.Panicf("%s, failed to connect to RabbitMQ", err)
	}
	defer pqconn.Close()

	mqllm, err := mqc.NewMQllm(qconn, pqconn, mqc.MQConfig{
		CancelExKey: mq_cancel_ex,
		ReqQKey:     mq_llm_q,
	})
	if err != nil {
		log.Panicf("%s, failed to initilize mq", err)
	}
	defer mqllm.Close()

	ctx := context.WithoutCancel(context.Background())

	completionsDone, err := mqllm.ConsumeCompletionsRequests(ctx, llm)
	if err != nil {
		log.Panicf("%s, failed to start consume completions", err)
	}
	cancellationsDone, err := mqllm.ConsumeCancellations(ctx, llm)
	if err != nil {
		log.Panicf("%s, failed to start consume cancellations", err)
	}

	<-completionsDone
	<-cancellationsDone
}

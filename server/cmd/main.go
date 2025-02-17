package main

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/soulnvkz/log"
	"github.com/soulnvkz/mq"
	mqc "github.com/soulnvkz/server/internal/mq"
	wsc "github.com/soulnvkz/server/internal/ws"
)

func Getenv(env string) (v string) {
	v, f := os.LookupEnv(env)
	if !f {
		log.Error().Fatalf("ENV %s should be specifed", env)
	}
	return
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (w *wrappedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		log.Info().Println(wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	log.Info().Print("Hello, server!")

	mq_user := Getenv("MQ_USER")
	mq_password := Getenv("MQ_PASSWORD")
	mq_host := Getenv("MQ_HOST")
	mq_port := Getenv("MQ_PORT")

	qconn, err := mq.MQConnect(mq_user, mq_password, mq_host, mq_port, 20)
	if err != nil {
		log.Error().Panicf("%s, failed to connect to RabbitMQ", err)
	}
	defer qconn.Close()

	pqconn, err := mq.MQConnect(mq_user, mq_password, mq_host, mq_port, 20)
	if err != nil {
		log.Error().Panicf("%s, failed to connect to RabbitMQ", err)
	}
	defer pqconn.Close()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	router := http.NewServeMux()
	router.HandleFunc("/completions", func(w http.ResponseWriter, r *http.Request) {
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error().Print(err)
		}

		websocket.SetCloseHandler(func(code int, text string) error {
			log.Info().Printf("closing ws connection. Code: %d, text:%s", code, text)
			return nil
		})

		mqcompeltions, err := mqc.NewMQCompletions(qconn, pqconn)
		if err != nil {
			log.Error().Print(err)
		}
		defer mqcompeltions.Close()

		socket := wsc.NewWSCompletions(r.Context(), websocket, mqcompeltions)
		defer socket.Close()

		socket.HandleMessages()
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: Logging(router),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Error().Fatal(err)
	}
}

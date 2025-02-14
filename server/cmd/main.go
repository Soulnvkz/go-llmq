package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/soulnvkz/server/internal/ws"
)

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

		log.Println(wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	log.Print("Hello, server!")
	time.Sleep(10 * time.Second)

	conn, err := amqp.Dial("amqp://admin:admin@rabbitmq:5672/")
	if err != nil {
		log.Panicf("%s: %s", "failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	pubConn, err := amqp.Dial("amqp://admin:admin@rabbitmq:5672/")
	if err != nil {
		log.Panicf("%s: %s", "failed to connect to RabbitMQ", err)
	}
	defer pubConn.Close()

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
			log.Println(err)
		}

		websocket.SetCloseHandler(func(code int, text string) error {
			log.Printf("Closing ws connection. Code: %d, text:%s", code, text)
			return nil
		})

		socket := ws.NewSocket(websocket, conn, pubConn, r.Context())
		socket.InitilizeRabbit()
		defer socket.Close()
		socket.HandleMessages()
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: Logging(router),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

func handleMessage(buff []byte, c *websocket.Conn) {
	message := string(buff)
	if message == "ping" {
		err := c.WriteMessage(websocket.TextMessage, []byte("pong"))
		if err != nil {
			log.Println("writing message error:", err)
		}
		return
	}

	r := strings.NewReader(fmt.Sprintf("answer to: %s", message))
	err := c.WriteMessage(websocket.TextMessage, []byte(`<start>`))
	if err != nil {
		log.Println("writing message error:", err)
	}
	for {
		abuff := make([]byte, 1)
		_, err := r.Read(abuff)
		if err == io.EOF {
			break
		}
		if len(abuff) > 0 {
			err = c.WriteMessage(websocket.TextMessage, abuff)
			if err != nil {
				log.Println("writing message error:", err)
				break
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
	err = c.WriteMessage(websocket.TextMessage, []byte(`<end>`))
	if err != nil {
		log.Println("writing message error:", err)
	}
}

// func handleMessages(ctxCancel context.CancelFunc, c *websocket.Conn) {
// 	ticker := time.NewTicker(1500 * time.Millisecond)
// 	defer ticker.Stop()
// 	gotPong := make(chan bool)

// 	go func() {
// 		for {
// 			select {
// 			case pong := <-gotPong:
// 				if pong {
// 					ticker.Reset(1500 * time.Millisecond)
// 					gotPong <- false
// 				}
// 			case <-ticker.C:
// 				ctxCancel()
// 			}
// 		}
// 	}()

// 	for {
// 		mt, buff, err := c.ReadMessage()
// 		if err != nil {
// 			log.Println("reading message error:", err)
// 			ctxCancel()
// 		}

// 		go handleMessage(buff, c, gotPong)

// 		_ = mt
// 		_ = buff
// 	}
// }

type Socket struct {
	c *websocket.Conn
}

func main() {
	log.Print("Hello, server!")

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	router := http.NewServeMux()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, test!"))
	})
	router.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, test1!"))
	})
	router.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
		}
		defer conn.Close()

		conn.SetCloseHandler(func(code int, text string) error {
			log.Printf("Closing ws connection. Code: %d, text:%s", code, text)
			return nil
		})

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for {
				mt, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("reading message error:", err)
					cancel()
					break
				}
				if mt == websocket.CloseMessage {
					log.Print("requesting close connection...")
					cancel()
					break
				}

				log.Printf("got message: %s", string(message))

				go handleMessage(message, conn)
			}
		}()

		<-ctx.Done()
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: Logging(router),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

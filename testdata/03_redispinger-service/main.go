package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	pingredis "github.com/bloomberg/docket/testdata/02_ping-redis"
)

func handler(w http.ResponseWriter, r *http.Request) {
	redisAddr := r.URL.Query().Get("redisAddr")
	if redisAddr == "" {
		w.WriteHeader(400)
		w.Write([]byte("missing redisAddr query parameter\n"))
		return
	}

	pong, err := pingredis.PingRedis(redisAddr)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("failed to ping redis: %v\n", err)))
		return
	}

	w.Write([]byte(pong))
	w.Write([]byte("\n"))
}

func main() {
	listenAddr := "localhost:0"
	if len(os.Args) >= 2 {
		listenAddr = os.Args[1]
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("err from net.Listen: %v", err)
	}

	fmt.Printf("Listening on %s\n", ln.Addr())

	http.HandleFunc("/", handler)

	if err := http.Serve(ln, nil); err != nil {
		log.Fatalf("err from http.Serve: %v", err)
	}
}

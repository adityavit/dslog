package server

import "net/http"

func NewHttpServer(addr string) *http.Server {
	httpServer := newHttpServer()
}

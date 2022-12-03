package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewHttpServer(addr string) *http.Server {
	r := mux.Router()
	return &http.Server{}
}

package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	Log *Log
}

type RecordData struct {
	Record Record `json:"record"`
}

type OffsetData struct {
	Offset uint64 `json:"offset"`
}

type ProduceRequest struct {
	RecordData
}

type ProduceResponse struct {
	OffsetData
}

type ConsumeRequest struct {
	OffsetData
}

type ConsumeResponse struct {
	RecordData
}

func NewHttpServer(addr string) *http.Server {
	httpServ := newHandleServer()
	r := mux.NewRouter()
	r.HandleFunc("/", httpServ.handleProduce).Methods("POST")
	r.HandleFunc("/", httpServ.handleConsume).Methods("GET")
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func newHandleServer() *Server {
	return &Server{
		Log: NewLog(),
	}
}

func (s *Server) handleProduce(w http.ResponseWriter, req *http.Request) {
	var pReq ProduceRequest
	err := json.NewDecoder(req.Body).Decode(&pReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	off, err := s.Log.Append(pReq.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pRes := ProduceResponse{OffsetData{Offset: off}}
	err = json.NewEncoder(w).Encode(pRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleConsume(w http.ResponseWriter, req *http.Request) {
	var cReq ConsumeRequest
	err := json.NewDecoder(req.Body).Decode(&cReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record, err := s.Log.Read(cReq.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ConsumeResponse{RecordData{record}}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

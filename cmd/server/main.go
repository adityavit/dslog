package main

import (
	"github.com/adityavit/dslog/internal/server"
	"log"
)

func main() {
	ser := server.NewHttpServer(":8080")
	log.Println("Stating http server @", ser.Addr)
	err := ser.ListenAndServe()
	if err != nil {
		log.Fatal("Unable to start server!")
	}
}

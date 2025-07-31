package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/enthus-golang/gowsdl/example/server/gen"
	"github.com/enthus-golang/gowsdl/soap"
)

var done = make(chan struct{})

func client() {
	client := soap.NewClient("http://127.0.0.1:8000")
	service := gen.NewMNBArfolyamServiceType(client)
	resp, err := service.GetInfoSoap(context.Background(), &gen.GetInfo{
		Id: "shenfuqiang",
	})
	fmt.Println(resp.GetInfoResult, err)
	done <- struct{}{}
}

// use fixtures/test.wsdl
func main() {
	// TODO: Server endpoint generation not implemented yet
	// http.HandleFunc("/", gen.Endpoint)
	
	fmt.Println("Example client-only demo")
	go func() {
		time.Sleep(time.Second * 1)
		client()
	}()
	
	// Simple echo server for demo purposes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		_, err := w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetInfoResponse xmlns="http://www.mnb.hu/webservices/">
      <GetInfoResult>Demo response for shenfuqiang</GetInfoResult>
    </GetInfoResponse>
  </soap:Body>
</soap:Envelope>`))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})
	
	go func() {
		if err := http.ListenAndServe(":8000", nil); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	<-done
}

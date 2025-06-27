package main

import (
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
	resp, err := service.GetInfoSoap(&gen.GetInfo{
		Id: "shenfuqiang",
	})
	fmt.Println(resp.GetInfoResult, err)
	done <- struct{}{}
}

// use fixtures/test.wsdl
func main() {
	http.HandleFunc("/", gen.Endpoint)
	go func() {
		time.Sleep(time.Second * 1)
		client()
	}()
	go func() {
		if err := http.ListenAndServe(":8000", nil); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	<-done
}

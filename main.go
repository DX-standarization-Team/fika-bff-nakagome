package main

import (
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/router"
)

func main() {
	rtr := router.New()
	if err := http.ListenAndServe("0.0.0.0:8080", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

package main

import (
	// "flag"
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/router"
)

func main() {

	// var env string
	// flag.StringVar(&env, "env", "dev", "環境")
	// flag.Parse()
	// log.Printf("RUNNING env: %s", env)

	rtr := router.New()
	if err := http.ListenAndServe("0.0.0.0:8080", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

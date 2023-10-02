package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/router"
)

var runningEnv string

func init() {
	flag.StringVar(&runningEnv, "runningEnv", "dev", "Environment to use")
}
func main() {
	flag.Parse()
	log.Printf("RUNNING runningEnv: %s", runningEnv)

	rtr := router.New()
	if err := http.ListenAndServe("0.0.0.0:8080", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

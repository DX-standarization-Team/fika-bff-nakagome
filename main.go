package main

import (
	// "embed"
	// "flag"
	// "fmt"
	"log"
	"net/http"

	// "path"

	// "gopkg.in/yaml.v3"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/router"
)



func main() {
	
	rtr := router.New()
	if err := http.ListenAndServe("0.0.0.0:8080", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

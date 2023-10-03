package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/router"
)

//go:embed config/*.yaml
var content embed.FS

var runningEnv string

type Config struct {
	Credentials struct {
		Password string `yaml:"password"`
	}
}

func init() {
	// 実行環境を取得
	flag.StringVar(&runningEnv, "runningEnv", "qc", "Environment to use")
}

func main() {
	// 実行環境の読み取り
	flag.Parse()
	log.Printf("runningEnv: %s", runningEnv)
	// 設定ファイル読み取り
	content, err := content.Open(path.Join("config", fmt.Sprintf("%s.yaml", runningEnv)))
	if err != nil {
		log.Fatalf("Failed to open content. err: %v", err)
	}
	config := &Config{}
	decoder := yaml.NewDecoder(content)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to decode yaml. err: %v", err)
	}
	log.Printf("Application setting file password: %s", config.Credentials.Password)

	rtr := router.New()
	if err := http.ListenAndServe("0.0.0.0:8080", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

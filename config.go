package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"path"

	"gopkg.in/yaml.v3"
)

//go:embed config/*.yaml
var content embed.FS

var RunningEnv string

type Config struct {
	Credentials struct {
		// Password string `yaml:"password"`
		PORT string `yaml:"port"`
	}
	Auth0 struct {
		AUTH0_DOMAIN   string `yaml:"auth0Domain"`
		AUTH0_AUDIENCE string `yaml:"auth0Audience"`
	}
}

func (c *Config) GetConfig() *Config {

	// 実行環境を取得
	flag.StringVar(&RunningEnv, "runningEnv", "develop", "Environment to use")

	// 実行環境の読み取り
	flag.Parse()
	log.Printf("runningEnv: %s", RunningEnv)

	// 設定ファイル読み取り
	content, err := content.Open(path.Join("config", fmt.Sprintf("%s.yaml", RunningEnv)))
	if err != nil {
		log.Fatalf("Failed to open content. err: %v", err)
	}
	config := &Config{}
	decoder := yaml.NewDecoder(content)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to decode yaml. err: %v", err)
	}
	log.Printf("Application setting file password: %v", config.Credentials)

	return config
}

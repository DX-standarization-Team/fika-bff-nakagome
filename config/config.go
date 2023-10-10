package config

import (
	"embed"
	"flag"
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
)

//go:embed *.yaml
var content embed.FS

type Config struct {
	Credentials struct {
		PORT string `yaml:"port"`
	}
	Auth0 struct {
		AUTH0_DOMAIN   string `yaml:"AUTH0_DOMAIN"`
		AUTH0_AUDIENCE string `yaml:"AUTH0_AUDIENCE"`
	}
}

// flag変数宣言
var RunningEnv string

func init() {
	// 実行環境を取得
	flag.StringVar(&RunningEnv, "runningEnv", "dev", "Environment to use")
}

func GetConfig() *Config {
	// 実行環境の読み取り
	flag.Parse()
	log.Printf("RunningEnv: %s", RunningEnv)

	// 設定ファイル読み取り
	b, err := content.ReadFile(fmt.Sprintf("%s.yaml", RunningEnv))
	if err != nil {
		log.Fatalf("Failed to open content. err: %v", err)
	}
	config := &Config{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		log.Fatalf("Failed to unmarshal content. err: %v", err)
	}
	return config
}

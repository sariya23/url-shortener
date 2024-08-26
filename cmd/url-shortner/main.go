package main

import (
	"fmt"
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

// func main() {
// 	// TODO: init config: cleanenv

// 	// TODO: init logger: slog

// 	// TODO: init storage

// 	// TODO: init router

// 	// TODO: run server
// }

type ConfigDatabase struct {
	Port string `yaml:"port" env:"PORT" env-default:"5432"`
}

func main() {
	var cfg ConfigDatabase
	err := cleanenv.ReadConfig("config.yaml", &cfg)

	if err != nil {
		fmt.Println(err)
	}

	slog.Info(cfg)
}

// Package config provides helpers to parse config files.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func ConfigFromFile(cfg any) {
	for i, arg := range os.Args {
		if arg != "-c" && arg != "--config" {
			continue
		}

		if i+1 >= len(os.Args) {
			fmt.Println("missing config file")
			os.Exit(2)
		}

		parseConfig(cfg, os.Args[i+1])
	}

	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "CONFIG=") {
			continue
		}

		err := parseConfig(cfg, strings.TrimPrefix(v, "CONFIG="))
		if err != nil {
			os.Exit(2)
		}
	}
}

func parseConfig(cfg any, fileName string) error {
	b, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("failed to read config file: %s\n", err)
		return err
	}

	err = json.Unmarshal(b, cfg)
	if err != nil {
		fmt.Printf("failed to parse config file: %s", err)
		return err
	}

	return nil
}

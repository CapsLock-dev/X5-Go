package config

import "os"

func HTTPAddr() string {
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		return ":8080"
	}
	return address
}

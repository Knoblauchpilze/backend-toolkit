package server

import "time"

type Config struct {
	BasePath        string
	Port            uint16
	ShutdownTimeout time.Duration
}

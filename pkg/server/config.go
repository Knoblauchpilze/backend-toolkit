package server

import "time"

type Config struct {
	BasePath        string
	ShutdownTimeout time.Duration
}

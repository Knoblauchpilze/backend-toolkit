package db

type Config interface {
	ToConnectionString() string
}

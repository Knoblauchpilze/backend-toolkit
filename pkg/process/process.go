package process

type RunFunc func()
type InterruptFunc func() error

type Process struct {
	Run       RunFunc
	Interrupt InterruptFunc
}

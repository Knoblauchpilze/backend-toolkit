package process

type RunFunc func() error
type InterruptFunc func() error

type Process struct {
	Run       RunFunc
	Interrupt InterruptFunc
}

func (p Process) Valid() bool {
	return p.Run != nil && p.Interrupt != nil
}

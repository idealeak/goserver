package netlib

type ioService interface {
	start() error
	update()
	shutdown()
	dump()
}

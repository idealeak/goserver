package netlib

import "net"

type Acceptor interface {
	ioService
	GetSessionConfig() *SessionConfig
	Addr() net.Addr
}

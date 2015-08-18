package netlib

import "time"

const (
	ReconnectInterval time.Duration = 5 * time.Second
)

type Connector interface {
	ioService
	GetSessionConfig() *SessionConfig
}

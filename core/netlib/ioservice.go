package netlib

type SessionStats struct {
	Id           int
	GroupId      int
	RunningTime  int64
	SendedBytes  int64
	RecvedBytes  int64
	SendedPack   int64
	RecvedPack   int64
	PendSendPack int
	PendRecvPack int
	RemoteAddr   string
}

type ServiceStats struct {
	Id           int
	Type         int
	Name         string
	Addr         string
	MaxActive    int
	MaxDone      int
	RunningTime  int64
	SessionStats []SessionStats
}

type ioService interface {
	start() error
	update()
	shutdown()
	dump()
	stats() ServiceStats
}

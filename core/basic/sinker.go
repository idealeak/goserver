package basic

type Sinker interface {
	OnStart()
	OnTick()
	OnStop()
}

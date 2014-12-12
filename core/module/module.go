package module

const (
	ModuleName_Net      string = "net-module"
	ModuleName_Transact        = "dtc-module"
)

type Module interface {
	ModuleName() string
	Init()
	Update()
	Shutdown()
}

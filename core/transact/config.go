// config
package transact

import (
	"github.com/idealeak/goserver/core"
)

var Config = Configuration{}

type Configuration struct {
	TxSkeletonName string
	tcs            TransactCommSkeleton
}

func (this *Configuration) Name() string {
	return "tx"
}

func (this *Configuration) Init() error {
	if this.TxSkeletonName != "" {
		this.tcs = GetTxCommSkeleton(this.TxSkeletonName)
	}
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}

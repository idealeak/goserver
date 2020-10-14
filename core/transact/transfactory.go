package transact

import (
	"fmt"

	"github.com/idealeak/goserver/core/logger"
)

var transactionHandlerPool = make(map[TransType]TransHandler)

func GetHandler(tt TransType) TransHandler {
	if v, exist := transactionHandlerPool[tt]; exist {
		return v
	}
	return nil
}

func RegisteHandler(tt TransType, th TransHandler) {
	if _, exist := transactionHandlerPool[tt]; exist {
		panic(fmt.Sprintf("TransHandlerFactory repeate registe handler, type=%v", tt))
		return
	}
	logger.Logger.Trace("transact.RegisteHandler:", tt)
	transactionHandlerPool[tt] = th
}

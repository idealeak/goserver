// config
package logger

import (
	"github.com/cihub/seelog"
)

var (
	Logger seelog.LoggerInterface
)

func init() {
	Logger, _ = seelog.LoggerFromConfigAsFile("logger.xml")
}

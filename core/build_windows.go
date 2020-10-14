package core

import (
	"fmt"
	"time"
)

func buildTime() string {
	return fmt.Sprintf("[%s]", time.Now().String())
}

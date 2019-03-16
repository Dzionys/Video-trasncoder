package lp

import (
	"fmt"

	"../sse"
)

func WLog(msg string) {
	fmt.Println(msg)
	sse.UpdateLogMessage(msg)
}

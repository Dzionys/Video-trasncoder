package lp

import (
	"fmt"

	"../sse"
)

func WLog(msg string) {
	fmt.Println(msg)
	go sse.UpdateLogMessage(msg)
}

package main

import (
	"os"
	"runtime"

	"golang.org/x/crypto/ssh/terminal"
)

const VERSION = "v0.1.0"

func main() {
	if len(os.Args) == 1 && (terminal.IsTerminal(0) || runtime.GOOS == "windows") {
		PromotMode()
		return
	}
	cmdRoot.Execute()
}

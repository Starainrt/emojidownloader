package main

import (
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

var version = "v0.1.0"

func main() {
	if len(os.Args) == 1 && terminal.IsTerminal(0) {
		PromotMode()
		return
	}
	cmdRoot.Execute()
}

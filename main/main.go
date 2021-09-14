package main

import (
	"os"
)

func main() {
	if len(os.Args) > 1 {
		RunFile(os.Args[1])
	} else {
		StartInteractiveShell()
	}
}

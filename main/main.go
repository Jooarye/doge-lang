package main

import (
	"doge/repl"
	"os"
)

func main() {
	repl.Start(os.Stdin, os.Stdin)
}

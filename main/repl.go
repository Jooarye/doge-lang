package main

import (
	"bufio"
	"doge/evaluator"
	"doge/lexer"
	"doge/object"
	"doge/parser"
	"fmt"
	"io"
)

var CommitId string
var DogeVersion = "0.1-dev"

var DOGE string = fmt.Sprintf("Doge v%s (commit: %s)", DogeVersion, CommitId)

const PROMPT = ">>> "

var history []string

func Start(in io.Reader, out io.Writer) {
	fmt.Println(DOGE)

	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	env.Set("__name__", &object.String{Value: "__main__"})
	evaluator.InitBuiltins()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			fmt.Println("Whoops such errors. Wow!!")
			fmt.Println("Syntax Errors:")
			PrintParserErrors(p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil && evaluated.Type() != object.NULL_OBJ {
			fmt.Println(evaluated.Inspect())
		}

		history = append(history, line)
	}
}

func PrintParserErrors(errors []string) {
	for _, msg := range errors {
		fmt.Printf("  %s\n", msg)
	}
}

package main

import (
	"doge/evaluator"
	"doge/lexer"
	"doge/object"
	"doge/parser"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		buf, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Couldn't read file! Aborting")
			os.Exit(0)
		}

		data := string(buf)
		l := lexer.New(data)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			fmt.Println("Whoops such errors. Wow!!")
			fmt.Println("Syntax Errors:")
			PrintParserErrors(p.Errors())
		}

		env := object.NewEnvironment()
		env.Set("__name__", &object.String{Value: "__main__"})
		evaluator.InitBuiltins()
		_ = evaluator.Eval(program, env)
	} else {
		Start(os.Stdin, os.Stdin)
	}
}

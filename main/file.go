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

func RunFile(path string) {
	buf, err := ioutil.ReadFile(path)
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
}

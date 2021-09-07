package repl

import (
	"bufio"
	"doge/evaluator"
	"doge/lexer"
	"doge/object"
	"doge/parser"
	"fmt"
	"io"
)

const DOGE = `          ▄              ▄
         ▌▒█           ▄▀▒▌
         ▌▒▒█        ▄▀▒▒▒▐
        ▐▄█▒▒▀▀▀▀▄▄▄▀▒▒▒▒▒▐
      ▄▄▀▒▒▒▒▒▒▒▒▒▒▒█▒▒▄█▒▐           __            _
    ▄▀▒▒▒░░░▒▒▒░░░▒▒▒▀██▀▒▌          / _\_   _  ___| |__
   ▐▒▒▒▄▄▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▀▄▒▌         \ \| | | |/ __| '_ \
   ▌░░▌█▀▒▒▒▒▒▄▀█▄▒▒▒▒▒▒▒█▒▐         _\ \ |_| | (__| | | |
  ▐░░░▒▒▒▒▒▒▒▒▌██▀▒▒░░░▒▒▒▀▄▌        \__/\__,_|\___|_| |_|
  ▌░▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░░░▒▒▒▒▌    
 ▌▒▒▒▄██▄▒▒▒▒▒▒▒▒░░░░░░░░▒▒▒▐         __
 ▐▒▒▐▄█▄█▌▒▒▒▒▒▒▒▒▒▒░▒░▒░▒▒▒▒▌       / /  __ _ _ __   __ _ _   _  __ _  __ _  ___
 ▐▒▒▐▀▐▀▒▒▒▒▒▒▒▒▒▒▒▒▒░▒░▒░▒▒▐       / /  / _` + "`" + ` | '_ \ / _` + "`" + ` | | | |/ _` + "`" + ` |/ _` + "`" + ` |/ _ \
  ▌▒▒▀▄▄▄▄▄▄▒▒▒▒▒▒▒▒░▒░▒░▒▒▒▌      / /__| (_| | | | | (_| | |_| | (_| | (_| |  __/
  ▐▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░▒░▒▒▄▒▒▐       \____/\__,_|_| |_|\__, |\__,_|\__,_|\__, |\___|
   ▀▄▒▒▒▒▒▒▒▒▒▒▒▒▒░▒░▒▄▒▒▒▒▌                         |___/             |___/
     ▀▄▒▒▒▒▒▒▒▒▒▒▄▄▄▀▒▒▒▒▄▀                      
       ▀▄▄▄▄▄▄▀▀▀▒▒▒▒▒▄▄▀
          ▀▀▀▀▀▀▀▀▀▀▀▀
`

const PROMPT = ">>> "

var history []string

func Start(in io.Reader, out io.Writer) {
	fmt.Println(DOGE)

	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

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

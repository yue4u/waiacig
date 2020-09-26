package repl

import (
	"bufio"
	"flag"
	"fmt"
	"io"

	"waiig/evaluator"
	"waiig/lexer"
	"waiig/object"
	"waiig/parser"
)

const MONKEY_FACE = `
            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

const PROMPT = ">> "

func StartREPL(in io.Reader, out io.Writer) {
	flag.Parse()
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()
	io.WriteString(out, MONKEY_FACE)
	for {
		fmt.Printf(PROMPT)
		if !scanner.Scan() {
			return
		}
		line := scanner.Text()
		l := lexer.NewLexer(line)
		p := parser.NewParser(l)
		program := p.ParseProgram()
		errors := p.Errors()
		if len(errors) != 0 {
			for _, msg := range errors {
				io.WriteString(out, "\t"+msg+"\n")
			}
			continue
		}
		evaluator.DefineMacros(program, macroEnv)
		expanded := evaluator.ExpandMacros(program, macroEnv)
		if evaluated := evaluator.Eval(expanded, env); evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

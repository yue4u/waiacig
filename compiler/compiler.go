package compiler

import (
	"waiacig/ast"
	"waiacig/code"
	"waiacig/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{}}
}

func (c *Compiler) Compile(node ast.Node) error {
	return nil
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

package compiler

import (
	"fmt"
	"testing"
	"waiacig/ast"
	"waiacig/code"
	"waiacig/lexer"
	"waiacig/object"
	"waiacig/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)
		compiler := NewCompiler()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	return p.ParseProgram()
}

func testInstructions(
	expected []code.Instructions,
	actual code.Instructions,
) error {
	concatted := concatInstructions(expected)
	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q",
			concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q", i, concatted, actual)
		}
	}
	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func testConstants(
	t *testing.T,
	expected []interface{},
	actual []object.Object,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s",
					i, err)
			}
		case []code.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T",
					i, actual[i])
			}
			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s",
					i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}
	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)",
			actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q",
			result.Value, expected)
	}
	return nil
}
func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpPop),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpSub),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpMul),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "2 / 1",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpDiv),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpPop)},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpFalse),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpGreaterThan),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpGreaterThan),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpNotEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpFalse),
				code.MakeInstruction(code.OpEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpFalse),
				code.MakeInstruction(code.OpNotEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpMinus),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpBang),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			if (true) {10}; 3333;
			`,
			expectedConstants: []interface{}{10, 3333},
			expectedInstructions: []code.Instructions{
				// 0000
				code.MakeInstruction(code.OpTrue),
				// 0001
				code.MakeInstruction(code.OpJumpNotTruthy, 10),
				// 0004
				code.MakeInstruction(code.OpConstant, 0),
				// 0007
				code.MakeInstruction(code.OpJump, 11),
				// 0010
				code.MakeInstruction(code.OpNull),
				// 0011
				code.MakeInstruction(code.OpPop),
				// 0012
				code.MakeInstruction(code.OpConstant, 1),
				// 0015
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        code.Opcode
		operands  []int
		bytesRead int
	}{
		{code.OpConstant, []int{65535}, 2},
	}
	for _, tt := range tests {
		instruction := code.MakeInstruction(tt.op, tt.operands...)
		def, err := code.Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}
		operandsRead, n := code.ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)
		}
		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
	let one = 1;
	let two = 2;
	`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpSetGlobal, 1),
			},
		},
		{
			input: `
	let one = 1;
	one;
	`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpGetGlobal, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input: `
let one = 1;
let two = one;
two;
`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpGetGlobal, 0),
				code.MakeInstruction(code.OpSetGlobal, 1),
				code.MakeInstruction(code.OpGetGlobal, 1),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []interface{}{"monkey"},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []interface{}{"mon", "key"},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpArray, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpArray, 3),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpConstant, 3),
				code.MakeInstruction(code.OpSub),
				code.MakeInstruction(code.OpConstant, 4),
				code.MakeInstruction(code.OpConstant, 5),
				code.MakeInstruction(code.OpMul),
				code.MakeInstruction(code.OpArray, 3),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpHash, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpConstant, 3),
				code.MakeInstruction(code.OpConstant, 4),
				code.MakeInstruction(code.OpConstant, 5),
				code.MakeInstruction(code.OpHash, 6),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpConstant, 3),
				code.MakeInstruction(code.OpConstant, 4),
				code.MakeInstruction(code.OpConstant, 5),
				code.MakeInstruction(code.OpMul),
				code.MakeInstruction(code.OpHash, 4),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpArray, 3),
				code.MakeInstruction(code.OpConstant, 3),
				code.MakeInstruction(code.OpConstant, 4),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpIndex),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpHash, 2),
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpConstant, 3),
				code.MakeInstruction(code.OpSub),
				code.MakeInstruction(code.OpIndex),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { return 5 + 10 }`,
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.MakeInstruction(code.OpConstant, 0),
					code.MakeInstruction(code.OpConstant, 1),
					code.MakeInstruction(code.OpAdd),
					code.MakeInstruction(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { }`,
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.MakeInstruction(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { 24 }();`,
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.MakeInstruction(code.OpConstant, 0), // The literal "24"
					code.MakeInstruction(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 1), // The compiled function
				code.MakeInstruction(code.OpCall),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input: `
	let noArg = fn() { 24 };
	noArg();
	`,
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.MakeInstruction(code.OpConstant, 0), // The literal "24"
					code.MakeInstruction(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 1), // The compiled function
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpGetGlobal, 0),
				code.MakeInstruction(code.OpCall),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
		let num = 55;
		fn() { num }
		`,
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.MakeInstruction(code.OpGetGlobal, 0),
					code.MakeInstruction(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input: `
		fn() {
		let num = 55;
		num
		}
		`,
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.MakeInstruction(code.OpConstant, 0),
					code.MakeInstruction(code.OpSetLocal, 0),
					code.MakeInstruction(code.OpGetLocal, 0),
					code.MakeInstruction(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input: `
		fn() {
		let a = 55;
		let b = 77;
		a + b
		}
		`,
			expectedConstants: []interface{}{
				55,
				77,
				[]code.Instructions{
					code.MakeInstruction(code.OpConstant, 0),
					code.MakeInstruction(code.OpSetLocal, 0),
					code.MakeInstruction(code.OpConstant, 1),
					code.MakeInstruction(code.OpSetLocal, 1),
					code.MakeInstruction(code.OpGetLocal, 0),
					code.MakeInstruction(code.OpGetLocal, 1),
					code.MakeInstruction(code.OpAdd),
					code.MakeInstruction(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 2),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := NewCompiler()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}
	globalSymbolTable := compiler.symbolTable
	compiler.emit(code.OpMul)
	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}
	compiler.emit(code.OpSub)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}
	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpSub)
	}
	if compiler.symbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}
	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d",
			compiler.scopeIndex, 0)
	}
	if compiler.symbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}
	compiler.emit(code.OpAdd)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}
	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpAdd)
	}
	previous := compiler.scopes[compiler.scopeIndex].previousInstruction
	if previous.Opcode != code.OpMul {
		t.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d",
			previous.Opcode, code.OpMul)
	}
}

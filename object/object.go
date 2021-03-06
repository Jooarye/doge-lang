package object

import (
	"bytes"
	"doge/ast"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	BREAK_OBJ        = "BREAK"
	NULL_OBJ         = "NULL"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType {
	return FLOAT_OBJ
}
func (f *Float) Inspect() string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

type Null struct{}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}
func (n *Null) Inspect() string {
	return "null"
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

type Break struct {
}

func (br *Break) Type() ObjectType {
	return BREAK_OBJ
}
func (br *Break) Inspect() string {
	return "BREAK"
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}
func (e *Error) Inspect() string {
	return fmt.Sprintf("ERROR: %s", e.Message)
}

type Environment struct {
	store map[string]Object
	outer *Environment
	loop  bool
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}
func (e *Environment) Set(name string, val Object) Object {
	if e.outer != nil {
		if _, ok := e.outer.Get(name); ok && e.loop {
			e.outer.store[name] = val
		}
	}

	e.store[name] = val
	return val
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil, loop: false}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewPartiallyEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnclosedEnvironment(outer)
	env.loop = true
	return env
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
}

func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("func")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}
func (s *String) Inspect() string {
	return s.Value
}

type BuiltinFunction func(env *Environment, args ...Object) Object

type Builtin struct {
	Fn            BuiltinFunction
	Documentation string
}

func (b *Builtin) Type() ObjectType {
	return BUILTIN_OBJ
}
func (b *Builtin) Inspect() string {
	return "builtin function"
}

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType {
	return ARRAY_OBJ
}
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType {
	return HASH_OBJ
}
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

package evaluator

import (
	"doge/ast"
	"doge/object"
	"fmt"
	"math"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return EvalProgram(node, env)
	case *ast.BlockStatement:
		return EvalBlockStatements(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if IsError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return EvalIdentifier(node, env)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if IsError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}

		return EvalInfixExpression(node.Operator, left, right)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}
		return EvalPrefixExpression(node.Operator, right)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.IfExpression:
		return EvalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if IsError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if IsError(function) {
			return function
		}
		args := EvalExpressions(node.Arguments, env)
		if len(args) == 1 && IsError(args[0]) {
			return args[0]
		}

		return ApplyFunction(function, args)
	case *ast.HashLiteral:
		return EvalHashLiteral(node, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return NativeBoolToBooleanObject(node.Value)
	case *ast.ArrayLiteral:
		elements := EvalExpressions(node.Elements, env)
		if len(elements) == 1 && IsError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if IsError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if IsError(index) {
			return index
		}

		return EvalIndexExpression(left, index)
	}

	return nil
}

func EvalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if IsError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return NewError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if IsError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func EvalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return EvalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return EvalHashIndexExpression(left, index)
	default:
		return NewError("index operator not supported: %s", left.Type())
	}
}

func EvalHashIndexExpression(left, index object.Object) object.Object {
	hashObj := left.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return NewError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func EvalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)

	if idx < 0 {
		idx = max + idx + 1
	}

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObj.Elements[idx]
}

func EvalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if IsError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func ApplyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		return RunFunction(fn, args)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return NewError("not a function: %s", fn.Type())
	}
}

func RunFunction(fn *object.Function, args []object.Object) object.Object {
	extendedEnv := ExtendFunctionEnv(fn, args)
	evaluated := Eval(fn.Body, extendedEnv)
	return UnwrapReturnValue(evaluated)
}

func ExtendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func UnwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func EvalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range program.Statements {
		result = Eval(stmt, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func EvalBlockStatements(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Eval(stmt, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func EvalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func EvalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return EvalBangOperatorExpression(right)
	case "-":
		return EvalMinusOperatorExpression(right)
	default:
		return NewError("unknown operator: %s%s", operator, right.Type())
	}
}

func EvalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return EvalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return EvalStringInfixExpression(operator, left, right)
	case operator == "==":
		return NativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return NativeBoolToBooleanObject(left != right)
	case operator == "&&":
		leftVal := left == TRUE
		rightVal := right == TRUE
		return NativeBoolToBooleanObject(leftVal && rightVal)
	case operator == "||":
		leftVal := left == TRUE
		rightVal := right == TRUE
		return NativeBoolToBooleanObject(leftVal || rightVal)
	case left.Type() != right.Type():
		return NewError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func EvalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "^":
		return &object.Integer{Value: leftVal ^ rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "**":
		return &object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "&":
		return &object.Integer{Value: leftVal & rightVal}
	case "|":
		return &object.Integer{Value: leftVal | rightVal}
	case "<<":
		return &object.Integer{Value: leftVal << rightVal}
	case ">>":
		return &object.Integer{Value: leftVal >> rightVal}
	case "&&":
		return NativeBoolToBooleanObject((leftVal > 0) && (rightVal > 0))
	case "||":
		return NativeBoolToBooleanObject((leftVal > 0) || (rightVal > 0))
	case "<":
		return NativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return NativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return NativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return NativeBoolToBooleanObject(leftVal >= rightVal)
	case "!=":
		return NativeBoolToBooleanObject(leftVal != rightVal)
	case "==":
		return NativeBoolToBooleanObject(leftVal == rightVal)
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func EvalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func EvalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if IsError(condition) {
		return condition
	}

	if IsTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func EvalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return NewError("identifier not found: %s", node.Value)
}

func EvalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func EvalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NewError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func NativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func IsTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func IsError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func NewError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

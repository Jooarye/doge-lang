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
	case *ast.AssignExpression:
		idt, ok := node.Left.(*ast.Identifier)
		if !ok {
			return NewError("cannot assign to non identifier!")
		}

		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}

		return EvalAssignExpression(node.TokenLiteral(), idt, right, env)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if IsError(right) {
			return right
		}
		return EvalPrefixExpression(node.Operator, right)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body}
	case *ast.IfExpression:
		return EvalIfExpression(node, env)
	case *ast.WhileExpression:
		return EvalWhileExpression(node, env)
	case *ast.ForExpression:
		return EvalForExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if IsError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.BreakStatement:
		return &object.Break{}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if IsError(function) {
			return function
		}

		args := EvalExpressions(node.Arguments, env)
		if len(args) == 1 && IsError(args[0]) {
			return args[0]
		}

		return ApplyFunction(function, args, env)
	case *ast.HashLiteral:
		return EvalHashLiteral(node, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
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

func EvalAssignExpression(literal string, idt *ast.Identifier, right object.Object, env *object.Environment) object.Object {
	val, ok := env.Get(idt.Value)

	result := right

	if literal != "=" {
		if !ok {
			return NewError("cannot assign to uninitialized identifier!")
		}

		if val.Type() != right.Type() {
			return NewError("cannot use %s with types: %s and %s", literal, val.Type(), right.Type())
		}

		if val.Type() == object.INTEGER_OBJ {
			lv := val.(*object.Integer)
			rv := right.(*object.Integer)
			switch literal {
			case "-=":
				result = &object.Integer{Value: lv.Value - rv.Value}
			case "+=":
				result = &object.Integer{Value: lv.Value + rv.Value}
			case "*=":
				result = &object.Integer{Value: lv.Value * rv.Value}
			case "/=":
				result = &object.Integer{Value: lv.Value / rv.Value}
			default:
				return NewError("Unknown assign operator %s", literal)
			}
		} else if val.Type() == object.FLOAT_OBJ {
			lv := val.(*object.Float)
			rv := right.(*object.Float)
			switch literal {
			case "-=":
				result = &object.Float{Value: lv.Value - rv.Value}
			case "+=":
				result = &object.Float{Value: lv.Value + rv.Value}
			case "*=":
				result = &object.Float{Value: lv.Value * rv.Value}
			case "/=":
				result = &object.Float{Value: lv.Value / rv.Value}
			default:
				return NewError("Unknown assign operator %s", literal)
			}
		} else if val.Type() == object.STRING_OBJ {
			lv := val.(*object.String)
			rv := right.(*object.String)
			if literal == "+=" {
				result = &object.String{Value: lv.Value + rv.Value}
			} else {
				return NewError("Unknown assign operator %s", literal)
			}
		}
	}

	env.Set(idt.Value, result)
	return NULL
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
	case left.Type() == object.STRING_OBJ:
		return EvalStringIndexExpression(left, index)
	default:
		return NewError("index operator not supported: %s", left.Type())
	}
}

func EvalStringIndexExpression(left object.Object, index object.Object) object.Object {
	strObj := left.(*object.String)
	idx, ok := index.(*object.Integer)

	if !ok {
		return NewError("Index for string can only be integer!")
	}

	if idx.Value > int64(len(strObj.Value)) || idx.Value < -int64(len(strObj.Value)) {
		return NewError("Index out of range!")
	}

	return &object.String{Value: fmt.Sprintf("%c", strObj.Value[idx.Value])}
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
		return NewError("index out of bounds")
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

func ApplyFunction(fn object.Object, args []object.Object, env *object.Environment) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		return RunFunction(fn, args)
	case *object.Builtin:
		return fn.Fn(env, args...)
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
	env := object.NewEnvironment()

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
			if rt == object.RETURN_VALUE_OBJ || rt == object.BREAK_OBJ || rt == object.ERROR_OBJ {
				return result
			}
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
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return EvalFloatInfixExpression(operator, left, right)
	case IsNumeric(left) && IsNumeric(right):
		return EvalMixedInfixExpression(operator, left, right)
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

func EvalMixedInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := ObjectToFloat(left)
	rightVal := ObjectToFloat(right)

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "%":
		return &object.Integer{Value: int64(leftVal) % int64(rightVal)}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
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
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
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

func EvalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "%":
		return &object.Integer{Value: int64(leftVal) % int64(rightVal)}
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
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "!=":
		return NativeBoolToBooleanObject(leftVal != rightVal)
	case "==":
		return NativeBoolToBooleanObject(leftVal == rightVal)
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func EvalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	pEnv := object.NewPartiallyEnclosedEnvironment(env)
	condition := Eval(ie.Condition, pEnv)
	if IsError(condition) {
		return condition
	}

	if IsTruthy(condition) {
		return Eval(ie.Consequence, pEnv)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, pEnv)
	} else {
		return NULL
	}
}

func EvalWhileExpression(we *ast.WhileExpression, env *object.Environment) object.Object {
	pEnv := object.NewPartiallyEnclosedEnvironment(env)
	var evaluated object.Object

	condition := Eval(we.Condition, pEnv)
	if IsError(condition) {
		return condition
	}

	for IsTruthy(condition) {
		evaluated = Eval(we.Consequence, pEnv)
		if IsError(evaluated) || evaluated.Type() == object.RETURN_VALUE_OBJ {
			return evaluated
		}

		if evaluated.Type() == object.BREAK_OBJ {
			return NULL
		}

		condition = Eval(we.Condition, pEnv)
		if IsError(condition) {
			return condition
		}
	}

	if evaluated != nil {
		return evaluated
	}

	return NULL
}

func EvalForExpression(fe *ast.ForExpression, env *object.Environment) object.Object {
	pEnv := object.NewEnclosedEnvironment(env)
	var evaluated object.Object

	initial := Eval(fe.Initial, pEnv)
	if IsError(initial) {
		return initial
	}

	condition := Eval(fe.Condition, pEnv)
	if IsError(condition) {
		return condition
	}

	for IsTruthy(condition) {
		evaluated = Eval(fe.Consequence, pEnv)
		if IsError(evaluated) || evaluated.Type() == object.RETURN_VALUE_OBJ {
			return evaluated
		}

		if evaluated.Type() == object.BREAK_OBJ {
			return NULL
		}

		increment := Eval(fe.Increment, pEnv)
		if IsError(increment) {
			return increment
		}

		condition = Eval(fe.Condition, pEnv)
		if IsError(condition) {
			return condition
		}
	}

	if evaluated != nil {
		return evaluated
	}

	return NULL
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
	if right.Type() != object.INTEGER_OBJ && right.Type() != object.FLOAT_OBJ {
		return NewError("unknown operator: -%s", right.Type())
	}

	if right.Type() == object.INTEGER_OBJ {
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	}

	value := right.(*object.Float).Value
	return &object.Float{Value: -value}
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

func ObjectToFloat(obj object.Object) float64 {
	switch obj := obj.(type) {
	case *object.Integer:
		return float64(obj.Value)
	case *object.Float:
		return obj.Value
	}

	return 0.0
}

func IsNumeric(obj object.Object) bool {
	return obj.Type() == object.FLOAT_OBJ || obj.Type() == object.INTEGER_OBJ
}

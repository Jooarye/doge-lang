package evaluator

import (
	"doge/object"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var builtins = map[string]*object.Builtin{}

func InitBuiltins() {
	builtins["append"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return NewError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to `append` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			arr.Elements = append(arr.Elements, args[1])

			return &object.Null{}
		},
	}
	builtins["remove"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return NewError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("argument to `remove` must be ARRAY, got %s", args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return NewError("second argument to `remove` must be INTEGER. got=%s", args[1].Type())
			}

			index := args[1].(*object.Integer)
			arr := args[0].(*object.Array)

			if int64(len(arr.Elements)) <= index.Value {
				return NewError("Index out of bounds!")
			}

			obj := arr.Elements[index.Value]
			arr.Elements = append(arr.Elements[:index.Value], arr.Elements[index.Value+1:]...)

			return obj
		},
	}
	builtins["print"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return NewError("print needs at least one argument. got=%d", len(args))
			}

			var elements []string
			for _, arg := range args {
				elements = append(elements, arg.Inspect())
			}

			out := strings.Join(elements, " ")

			fmt.Println(out)
			return &object.Null{}
		},
	}
	builtins["len"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Hash:
				return &object.Integer{Value: int64(len(arg.Pairs))}
			default:
				return NewError("argument to `len` not supported, got=%s", args[0].Type())
			}
		},
	}
	builtins["sum"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			value := int64(0)

			for _, elm := range array.Elements {
				if elm.Type() == object.INTEGER_OBJ {
					intObj := elm.(*object.Integer)
					value += intObj.Value
				}
			}

			return &object.Integer{Value: value}
		},
	}
	builtins["sumf"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			value := float64(0)

			for _, elm := range array.Elements {
				if elm.Type() == object.FLOAT_OBJ {
					floatObj := elm.(*object.Float)
					value += floatObj.Value
				}
			}

			return &object.Float{Value: value}
		},
	}
	builtins["min"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			value := int64(math.MaxInt64)

			for _, elm := range array.Elements {
				if elm.Type() == object.INTEGER_OBJ {
					intObj := elm.(*object.Integer)
					if intObj.Value < value {
						value = intObj.Value
					}
				}
			}

			return &object.Integer{Value: value}
		},
	}
	builtins["minf"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			value := math.MaxFloat64

			for _, elm := range array.Elements {
				if elm.Type() == object.FLOAT_OBJ {
					floatObj := elm.(*object.Float)
					if floatObj.Value < value {
						value = floatObj.Value
					}
				}
			}

			return &object.Float{Value: value}
		},
	}
	builtins["max"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			value := int64(0)

			for _, elm := range array.Elements {
				if elm.Type() == object.INTEGER_OBJ {
					intObj := elm.(*object.Integer)
					if intObj.Value > value {
						value = intObj.Value
					}
				}
			}

			return &object.Integer{Value: value}
		},
	}
	builtins["maxf"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			value := float64(0)

			for _, elm := range array.Elements {
				if elm.Type() == object.FLOAT_OBJ {
					floatObj := elm.(*object.Float)
					if floatObj.Value > value {
						value = floatObj.Value
					}
				}
			}

			return &object.Float{Value: value}
		},
	}
	builtins["int"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("expected 1 argument. got=%d", len(args))
			}

			if args[0].Type() == object.FLOAT_OBJ {
				floatObj, ok := args[0].(*object.Float)
				if !ok {
					return NewError("argument cannot be interpreted as FLOAT")
				}

				return &object.Integer{Value: int64(floatObj.Value)}
			} else if args[0].Type() == object.STRING_OBJ {
				stringObj, ok := args[0].(*object.String)
				if !ok {
					return NewError("argument cannot be interpreted as STRING")
				}
				val, err := strconv.ParseInt(stringObj.Value, 10, 64)
				if err != nil {
					return NewError("couldn't parse string as integer")
				}

				return &object.Integer{Value: val}
			}

			return NewError("argument to int must be string or float. got=%s", args[0].Type())
		},
	}
	builtins["float"] = &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("expected 1 argument. got=%d", len(args))
			}

			if args[0].Type() == object.INTEGER_OBJ {
				intObj, ok := args[0].(*object.Integer)
				if !ok {
					return NewError("argument cannot be interpreted as FLOAT")
				}

				return &object.Float{Value: float64(intObj.Value)}
			} else if args[0].Type() == object.STRING_OBJ {
				stringObj, ok := args[0].(*object.String)
				if !ok {
					return NewError("argument cannot be interpreted as STRING")
				}
				val, err := strconv.ParseFloat(stringObj.Value, 64)
				if err != nil {
					return NewError("couldn't parse string as float")
				}

				return &object.Float{Value: val}
			}

			return NewError("argument to int must be string or int. got=%s", args[0].Type())
		},
	}
	builtins["map"] = &object.Builtin{
		Fn: mapBuiltin,
	}
}

func mapBuiltin(args ...object.Object) object.Object {
	if len(args) != 2 {
		return NewError("wrong number of arguments. got=%d, want=2", len(args))
	}
	if args[0].Type() != object.ARRAY_OBJ {
		return NewError("argument to `map` must be ARRAY, got %s", args[0].Type())
	}
	if args[1].Type() != object.FUNCTION_OBJ && args[1].Type() != object.BUILTIN_OBJ {
		return NewError("second argument to `map` must be FUNCTION, got=%s", args[1].Type())
	}

	var elements []object.Object

	arr := args[0].(*object.Array)
	if args[1].Type() == object.FUNCTION_OBJ {
		funct := args[1].(*object.Function)

		for _, elm := range arr.Elements {
			res := RunFunction(funct, []object.Object{elm})
			elements = append(elements, res)
		}
	} else {
		bi := args[1].(*object.Builtin)

		for _, elm := range arr.Elements {
			res := bi.Fn(elm)
			elements = append(elements, res)
		}
	}

	return &object.Array{Elements: elements}
}

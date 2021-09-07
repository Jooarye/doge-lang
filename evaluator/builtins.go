package evaluator

import (
	"doge/object"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var builtins = map[string]*object.Builtin{
	"append": &object.Builtin{
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
	},
	"print": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return NewError("print needs at least one argument. got=%d", len(args))
			}

			elements := []string{}
			for _, arg := range args {
				elements = append(elements, arg.Inspect())
			}

			out := strings.Join(elements, " ")

			fmt.Println(out)
			return &object.Null{}
		},
	},
	"len": &object.Builtin{
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
	},
	"sum": &object.Builtin{
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
	},
	"min": &object.Builtin{
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
	},
	"max": &object.Builtin{
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
	},
	"int": &object.Builtin{
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
	},
	"float": &object.Builtin{
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
	},
}

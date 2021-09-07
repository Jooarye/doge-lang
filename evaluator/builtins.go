package evaluator

import (
	"doge/object"
	"fmt"
	"math"
	"strings"
)

var builtins = map[string]*object.Builtin{
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
				switch arg := arg.(type) {
				case *object.String:
					elements = append(elements, arg.Value)
				case *object.Integer:
					elements = append(elements, fmt.Sprintf("%d", arg.Value))
				case *object.Boolean:
					elements = append(elements, fmt.Sprintf("%t", arg.Value))
				default:
					return NewError("object cannot be represented as string. got=%T", arg)
				}
			}

			out := strings.Join(elements, " ")

			fmt.Println(out)
			return &object.Integer{Value: int64(len(out))}
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
}

package evaluator

import (
	"doge/lexer"
	"doge/object"
	"doge/parser"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
)

var builtins = map[string]*object.Builtin{}

func InitBuiltins() {
	builtins["append"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
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
		Documentation: "This function appends an object to a given array!",
	}
	builtins["remove"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
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
		Documentation: "This function removes on object from an array!",
	}
	builtins["print"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
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
		Documentation: "This function prints every object that is given to it, multiple arguments will be seperated by a space!",
	}
	builtins["len"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
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
		Documentation: "This function returns the length of an array, string or hash!",
	}
	builtins["sum"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			isFloat := false
			value := float64(0)

			for _, elm := range array.Elements {
				if elm.Type() == object.INTEGER_OBJ {
					intObj := elm.(*object.Integer)
					value += float64(intObj.Value)
				} else if elm.Type() == object.FLOAT_OBJ {
					fltObj := elm.(*object.Float)
					value += fltObj.Value
					isFloat = true
				}
			}

			if isFloat {
				return &object.Float{Value: value}
			}

			return &object.Integer{Value: int64(value)}
		},
		Documentation: "This function returns the sum of an array!",
	}
	builtins["min"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			isFloat := false
			value := float64(math.MaxFloat64)

			for _, elm := range array.Elements {
				if elm.Type() == object.INTEGER_OBJ {
					intObj := elm.(*object.Integer)

					if float64(intObj.Value) < value {
						value = float64(intObj.Value)
					}
				} else if elm.Type() == object.FLOAT_OBJ {
					fltObj := elm.(*object.Float)

					if fltObj.Value < value {
						isFloat = true
						value = fltObj.Value
					}
				}
			}

			if isFloat {
				return &object.Float{Value: value}
			}

			return &object.Integer{Value: int64(value)}
		},
		Documentation: "This function returns the smallest value of an array!",
	}
	builtins["max"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("argument to `sum` must be array.")
			}
			array, ok := args[0].(*object.Array)
			if !ok {
				return NewError("argument to `sum` must be array. got=%T", args[0])
			}

			isFloat := false
			value := float64(0)

			for _, elm := range array.Elements {
				if elm.Type() == object.INTEGER_OBJ {
					intObj := elm.(*object.Integer)
					if float64(intObj.Value) > value {
						value = float64(intObj.Value)
					}
				} else if elm.Type() == object.FLOAT_OBJ {
					fltObj := elm.(*object.Float)
					if fltObj.Value > value {
						value = fltObj.Value
						isFloat = true
					}
				}
			}

			if isFloat {
				return &object.Float{Value: value}
			}

			return &object.Integer{Value: int64(value)}
		},
		Documentation: "This function returns the max value of an array!",
	}
	builtins["int"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
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
		Documentation: "This function converts a string or float to an int!",
	}
	builtins["float"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
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
		Documentation: "This function converts a string or int to a float!",
	}
	builtins["map"] = &object.Builtin{
		Fn:            mapBuiltin,
		Documentation: "This function calls a function for every entry in an array and adds the result to a new one!",
	}
	builtins["import"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return NewError("import expected at least 1 argument. got=%d", len(args))
			}

			env.Set("__name__", &object.String{Value: "__import__"})
			for _, arg := range args {
				strObj, ok := arg.(*object.String)
				if !ok {
					env.Set("__name__", &object.String{Value: "__main__"})
					return NewError("argument to import must be string. got=%s", arg.Type())
				}

				filePath := fmt.Sprintf("%s.doge", strObj.Value)

				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					root, ok := os.LookupEnv("DOGEROOT")
					if !ok {
						return NewError("couldn't find file")
					}

					filePath = path.Join(root, filePath)
				}

				buf, err := ioutil.ReadFile(filePath)
				if err != nil {
					env.Set("__name__", &object.String{Value: "__main__"})
					return NewError("cannot import file '%s'", strObj.Value)
				}

				data := string(buf)
				l := lexer.New(data)
				p := parser.New(l)
				program := p.ParseProgram()

				if len(p.Errors()) != 0 {
					env.Set("__name__", &object.String{Value: "__main__"})
					return NewError("errors while importing file '%s'\n\t%s", strObj.Value, strings.Join(p.Errors(), "\n\t"))
				}

				_ = Eval(program, env)
			}

			env.Set("__name__", &object.String{Value: "__main__"})
			return &object.Null{}
		},
		Documentation: "This function imports other doge files!",
	}
	builtins["help"] = &object.Builtin{
		Fn:            helpBuiltin,
		Documentation: "Print this menu!",
	}
}

func helpBuiltin(env *object.Environment, args ...object.Object) object.Object {
	fmt.Println("Name\tDocumentation")
	fmt.Println("---------------------")
	for key, val := range builtins {
		fmt.Printf("%s\t%s\n", key, val.Documentation)
	}
	return &object.Null{}
}

func mapBuiltin(env *object.Environment, args ...object.Object) object.Object {
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
			res := bi.Fn(env, elm)
			elements = append(elements, res)
		}
	}

	return &object.Array{Elements: elements}
}

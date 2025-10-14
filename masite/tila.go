package masite

import "text/scanner"
import "bytes"
import "strconv"
import "fmt"

type TilaFunc = func(t *Tila, args ...any) any

type Tila struct {
	scanner.Scanner
	Env       map[Ident]any
	Commands  map[Ident]TilaFunc
	Operators map[Operator]TilaFunc
	In        *bytes.Buffer
	Out       *bytes.Buffer
	Err       *bytes.Buffer
}

func NewTila() *Tila {
	res := &Tila{}
	res.In = &bytes.Buffer{}
	res.Out = &bytes.Buffer{}
	res.Err = &bytes.Buffer{}
	res.Env = make(map[Ident]any)
	res.Commands = make(map[Ident]TilaFunc)
	res.Operators = make(map[Operator]TilaFunc)
	res.Scanner.Whitespace = 1<<'\t' | 1<<' '
	return res
}

type Ident string
type Operator string

func Value(token rune, text string) (res any) {
	switch token {
	case scanner.EOF:
		return nil
	case scanner.Comment:
		return nil
	case scanner.Int:
		v, err := strconv.Atoi(text)
		if err != nil {
			return err
		}
		return v
	case scanner.Float:
		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		return v
	case scanner.Ident:
		return Ident(text)
	case scanner.String:
		return
		v, err := strconv.Unquote(text)
		if err != nil {
			return err
		}
		return v

	case scanner.RawString:
		v, err := strconv.Unquote(text)
		if err != nil {
			return err
		}
		return v

	case '\n', '\r':
		return nil
	default:
		return Operator(text)
	}

}

func (t *Tila) Exec(line []any) any {
	if len(line) < 1 {
		return nil
	}
	name := line[0]
	if ident, ok := name.(Ident); ok {
		if cmd, ok := t.Commands[ident]; ok {
			return cmd(t, line...)
		} else {
			return fmt.Errorf("unknown command :%s", string(ident))
		}
	}
	if oper, ok := name.(Operator); ok {
		if cmd, ok := t.Operators[oper]; ok {
			return cmd(t, line...)
		} else {
			return fmt.Errorf("unknown operator :%s", string(oper))
		}
	}
	return name
}

func (t *Tila) Run(script string) any {
	t.Scanner.Init(bytes.NewBuffer([]byte(script)))
	t.Scanner.Whitespace = 1<<'\t' | 1<<' '
	line := []any{}
	all := []any{}
	minus := false
	for token := t.Scanner.Scan(); token != scanner.EOF; token = t.Scanner.Scan() {
		if token == '\n' || token == '\r' && len(line) > 0 {
			res := t.Exec(line)
			if err, isErr := res.(error); isErr {
				fmt.Fprintf(t.Err, "%s: %s\n", t.Scanner.Pos(), err)
				return err
			} else {
				all = append(all, res)
				fmt.Fprintf(t.Out, "%v\n", res)
			}
			line = []any{}
		} else if token == '-' {
			minus = true
		} else {
			val := Value(token, t.Scanner.TokenText())
			if minus {
				minus = false
				ival, ok := val.(int)
				if ok {
					line = append(line, -ival)
				} else {
					line = append(line, Operator("-"), val)
				}
			} else {
				line = append(line, val)
			}
		}
	}
	if len(line) > 0 {
		res := t.Exec(line)
		if err, isErr := res.(error); isErr {
			fmt.Fprintf(t.Err, "%s: %s\n", t.Scanner.Pos(), err)
			return err
		} else {
			all = append(all, res)
			fmt.Fprintf(t.Out, "%v\n", res)
		}
	}

	return all
}

func TilaArg[T any](args []any) (T, error) {
	var zero T
	if len(args) < 2 {
		return zero, fmt.Errorf("%s: needs 1 argument", args[0])
	}
	if val, ok := args[1].(T); ok {
		return val, nil
	} else {
		return zero, fmt.Errorf("%s: variable type not correct: %v expect %T",
			args[0], args[1], zero)
	}
}

func TilaArg2[T, U any](args []any) (T, U, error) {
	var zero1 T
	var zero2 U

	var r1 T
	var r2 U
	if len(args) < 3 {
		return zero1, zero2, fmt.Errorf("%s: needs 1 argument", args[0])
	}
	if val, ok := args[1].(T); ok {
		r1 = val
	} else {
		return zero1, zero2, fmt.Errorf("%s: variable type not correct: %v expect %T",
			args[0], args[1], zero1)
	}
	if val, ok := args[2].(U); ok {
		r2 = val
	} else {
		return zero1, zero2, fmt.Errorf("%s: variable type not correct: %v expect %T",
			args[0], args[1], zero2)
	}
	return r1, r2, nil
}

func (t *Tila) Set(args ...any) any {
	if ident, val, err := TilaArg2[Ident, any](args); err != nil {
		return err
	} else {
		t.Env[ident] = val
		return val
	}
}

func (t *Tila) Get(args ...any) any {
	if ident, err := TilaArg[Ident](args); err != nil {
		return err
	} else {
		if val, ok := t.Env[ident]; ok {
			return val
		} else {
			return fmt.Errorf("get: variable not set: %v", args[1])
		}
	}
}

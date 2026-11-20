// Calculator: lexer, recursive-descent parser, AST with an interface and an evaluator.
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type TokenKind int

const (
	NUM TokenKind = iota
	IDENT
	OP
	LPAREN
	RPAREN
	COMMA
	EOF
)

func (k TokenKind) String() string {
	return [...]string{"NUM", "IDENT", "OP", "LPAREN", "RPAREN", "COMMA", "EOF"}[k]
}

type Token struct {
	Kind TokenKind
	Text string
	Pos  int
}

func lex(src string) ([]Token, error) {
	var toks []Token
	rs := []rune(src)
	for i := 0; i < len(rs); {
		r := rs[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case unicode.IsDigit(r) || r == '.':
			start := i
			for i < len(rs) && (unicode.IsDigit(rs[i]) || rs[i] == '.') {
				i++
			}
			toks = append(toks, Token{NUM, string(rs[start:i]), start})
		case unicode.IsLetter(r):
			start := i
			for i < len(rs) && (unicode.IsLetter(rs[i]) || unicode.IsDigit(rs[i])) {
				i++
			}
			toks = append(toks, Token{IDENT, string(rs[start:i]), start})
		case strings.ContainsRune("+-*/^%", r):
			toks = append(toks, Token{OP, string(r), i})
			i++
		case r == '(':
			toks = append(toks, Token{LPAREN, "(", i})
			i++
		case r == ')':
			toks = append(toks, Token{RPAREN, ")", i})
			i++
		case r == ',':
			toks = append(toks, Token{COMMA, ",", i})
			i++
		default:
			return nil, fmt.Errorf("unexpected %q at %d", r, i)
		}
	}
	return append(toks, Token{EOF, "", len(rs)}), nil
}

type Node interface {
	Eval(env map[string]float64) (float64, error)
	String() string
}

type Num float64
type Var string
type Unary struct {
	Op string
	X  Node
}
type Binary struct {
	Op   string
	L, R Node
}
type Call struct {
	Name string
	Args []Node
}

func (n Num) Eval(map[string]float64) (float64, error) { return float64(n), nil }
func (n Num) String() string                           { return strconv.FormatFloat(float64(n), 'g', -1, 64) }
func (v Var) Eval(env map[string]float64) (float64, error) {
	if x, ok := env[string(v)]; ok {
		return x, nil
	}
	return 0, fmt.Errorf("undefined variable %s", string(v))
}
func (v Var) String() string { return string(v) }
func (u Unary) Eval(env map[string]float64) (float64, error) {
	x, err := u.X.Eval(env)
	return -x, err
}
func (u Unary) String() string { return "(" + u.Op + u.X.String() + ")" }
func (b Binary) Eval(env map[string]float64) (float64, error) {
	l, err := b.L.Eval(env)
	if err != nil {
		return 0, err
	}
	r, err := b.R.Eval(env)
	if err != nil {
		return 0, err
	}
	switch b.Op {
	case "+":
		return l + r, nil
	case "-":
		return l - r, nil
	case "*":
		return l * r, nil
	case "/":
		if r == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return l / r, nil
	case "%":
		return math.Mod(l, r), nil
	case "^":
		return math.Pow(l, r), nil
	}
	return 0, fmt.Errorf("bad operator %s", b.Op)
}
func (b Binary) String() string { return "(" + b.L.String() + " " + b.Op + " " + b.R.String() + ")" }

var funcs = map[string]func(args []float64) (float64, error){
	"sqrt": func(a []float64) (float64, error) { return math.Sqrt(a[0]), nil },
	"max": func(a []float64) (float64, error) {
		m := a[0]
		for _, x := range a[1:] {
			m = math.Max(m, x)
		}
		return m, nil
	},
	"abs": func(a []float64) (float64, error) { return math.Abs(a[0]), nil },
}

func (c Call) Eval(env map[string]float64) (float64, error) {
	f, ok := funcs[c.Name]
	if !ok {
		return 0, fmt.Errorf("unknown function %s", c.Name)
	}
	var args []float64
	for _, a := range c.Args {
		v, err := a.Eval(env)
		if err != nil {
			return 0, err
		}
		args = append(args, v)
	}
	if len(args) == 0 {
		return 0, fmt.Errorf("%s needs arguments", c.Name)
	}
	return f(args)
}
func (c Call) String() string {
	parts := make([]string, len(c.Args))
	for i, a := range c.Args {
		parts[i] = a.String()
	}
	return c.Name + "(" + strings.Join(parts, ", ") + ")"
}

type Parser struct {
	toks []Token
	pos  int
}

func (p *Parser) peek() Token { return p.toks[p.pos] }
func (p *Parser) next() Token { t := p.toks[p.pos]; p.pos++; return t }

var prec = map[string]int{"+": 1, "-": 1, "*": 2, "/": 2, "%": 2, "^": 3}

func (p *Parser) expr(minPrec int) (Node, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		pr, ok := prec[t.Text]
		if t.Kind != OP || !ok || pr < minPrec {
			return left, nil
		}
		p.next()
		nextMin := pr + 1
		if t.Text == "^" {
			nextMin = pr
		}
		right, err := p.expr(nextMin)
		if err != nil {
			return nil, err
		}
		left = Binary{t.Text, left, right}
	}
}

func (p *Parser) unary() (Node, error) {
	if t := p.peek(); t.Kind == OP && t.Text == "-" {
		p.next()
		x, err := p.unary()
		return Unary{"-", x}, err
	}
	return p.primary()
}

func (p *Parser) primary() (Node, error) {
	t := p.next()
	switch t.Kind {
	case NUM:
		f, err := strconv.ParseFloat(t.Text, 64)
		if err != nil {
			return nil, fmt.Errorf("bad number %q", t.Text)
		}
		return Num(f), nil
	case IDENT:
		if p.peek().Kind != LPAREN {
			return Var(t.Text), nil
		}
		p.next()
		call := Call{Name: t.Text}
		for p.peek().Kind != RPAREN {
			arg, err := p.expr(1)
			if err != nil {
				return nil, err
			}
			call.Args = append(call.Args, arg)
			if p.peek().Kind == COMMA {
				p.next()
			} else if p.peek().Kind != RPAREN {
				return nil, fmt.Errorf("expected , or ) at %d", p.peek().Pos)
			}
		}
		p.next()
		return call, nil
	case LPAREN:
		n, err := p.expr(1)
		if err != nil {
			return nil, err
		}
		if p.next().Kind != RPAREN {
			return nil, fmt.Errorf("missing )")
		}
		return n, nil
	}
	return nil, fmt.Errorf("unexpected %v %q at %d", t.Kind, t.Text, t.Pos)
}

func Calc(src string, env map[string]float64) (string, error) {
	toks, err := lex(src)
	if err != nil {
		return "", err
	}
	p := &Parser{toks: toks}
	tree, err := p.expr(1)
	if err != nil {
		return "", err
	}
	if p.peek().Kind != EOF {
		return "", fmt.Errorf("trailing %q at %d", p.peek().Text, p.peek().Pos)
	}
	v, err := tree.Eval(env)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s = %.6g", tree, v), nil
}

func main() {
	env := map[string]float64{"pi": math.Pi, "x": 4, "rate": 0.07}
	for _, src := range []string{
		"1 + 2 * 3",
		"(1 + 2) * 3",
		"2 ^ 3 ^ 2",
		"-x ^ 2 + 10",
		"sqrt(16) + max(1, x, 3) * 2",
		"100 * (1 + rate) ^ 10",
		"17 % 5 + abs(-2.5)",
		"2 * pi * x",
		"1 / (x - 4)",
		"3 + ",
		"y * 2",
		"foo(1)",
		"4 $ 4",
		"(1 + 2",
	} {
		out, err := Calc(src, env)
		if err != nil {
			fmt.Printf("%-30s error: %v\n", src, err)
			continue
		}
		fmt.Printf("%-30s %s\n", src, out)
	}
}

package formula

import (
	"fmt"
	"math"
	"strconv"
)

/*
```ebnf
expression = term, { ("+" | "-"), term };
term       = factor, { ("*" | "/"), factor };
factor     = number | variable | function | "(", expression, ")";
function   = identifier, "(", expression, { ",", expression }, ")";
variable   = identifier;
identifier = [A-Z]+[A-Z0-9]*; // Added identifier rule
number     = [0-9]+("."[0-9]+)?;
```

```ebnf
program        = { statement, ";" };
statement      = assignment | expression;
assignment     = identifier, assign_op, expression;
expression     = term, { ("+" | "-"), term };
term           = factor, { ("*" | "/"), factor };
factor         = number | variable | function_call | "(", expression, ")";
function_call  = identifier, "(", argument_list, ")";
argument_list  = expression, { ",", expression };
identifier     = [A-Z]+[A-Z0-9]*;
assign_op      = ":=" | ":";
number         = [0-9]+("."[0-9]+)?;
reserved_var   = "CLOSE" | "OPEN" | "HIGH" | "LOW" ; // 保留变量
reserved_func  = "MA" | "REF" | "HHV" | "LLV" | "SMA" | "WMA" | "EMA"; // 保留函数

```

在这个 EBNF 中：

* `program` 是一个语句序列。
* `statement` 可以是一个赋值语句或一个表达式。
* `assignment` 是一个变量赋值语句，包括变量名、赋值运算符和表达式。
* `expression`、`term` 和 `factor` 定义了表达式的语法规则，与之前类似，但现在包含了函数调用。
* `function_call` 定义了函数调用的语法规则，包括函数名和参数列表。
* `argument_list` 定义了函数参数列表的语法规则。
* `identifier` 定义了标识符的语法规则。
* `assign_op` 定义了赋值运算符。
* `number` 定义了数字的语法规则。
* `reserved_var` 定义了保留变量。
* `reserved_func` 定义了保留函数。
*/
type NodeType string

const (
	NUMBER_NODE     NodeType = "NUMBER"
	OPERATOR_NODE   NodeType = "OPERATOR"
	EXPRESSION_NODE NodeType = "EXPRESSION"
	VARIABLE_NODE   NodeType = "VARIABLE" // 新增变量节点类型
	SYMBOL_NODE     NodeType = "SYMBOL"   // 新增符号节点类型
	FUNCTION_NODE   NodeType = "FUNCTION"
)

type Node struct {
	Type     NodeType
	Value    string
	Children []*Node
	Result   []float64
}

type SymbolTable map[string][]float64

// Parser structure
type Parser struct {
	tokens        []Token
	cursor        int
	data          map[string][]float64
	symbolTable   SymbolTable
	reservedWords map[string]bool
}

func NewParser(tokens []Token, data map[string][]float64) *Parser {
	reservedWords := make(map[string]bool)
	reservedWords["CLOSE"] = true
	reservedWords["OPEN"] = true
	reservedWords["HIGH"] = true
	reservedWords["LOW"] = true
	reservedWords["MA"] = true
	reservedWords["REF"] = true
	reservedWords["HHV"] = true
	reservedWords["LLV"] = true
	reservedWords["SMA"] = true
	reservedWords["WMA"] = true
	reservedWords["EMA"] = true
	return &Parser{tokens: tokens, data: data, symbolTable: make(SymbolTable), reservedWords: reservedWords}
}

func (p *Parser) Result() SymbolTable {
	return p.symbolTable
}

func (p *Parser) ParseApp() error {
	for {
		token, err := p.nextToken()
		if err != nil {
			if err.Error() == "no more tokens" {
				return nil
			}
			return err
		}
		p.cursor--
		err = p.parseStatement()
		if err != nil {
			return err
		}
		token, err = p.nextToken()
		if err != nil {
			return err
		}
		if token.Type != SEMICOLON {
			return fmt.Errorf("expected ';'")
		}
	}
}

func (p *Parser) parseStatement() error {
	token, err := p.nextToken()
	if err != nil {
		return err
	}
	if token.Type == IDENTIFIER {
		next, err := p.nextToken()
		if err != nil {
			return err
		}
		if next.Type == ASSIGN_OP {
			p.cursor -= 2 // 回退两个游标，保留变量名token
			return p.parseAssignment()
		} else {
			p.cursor -= 2 // 回退两个游标
			_, err := p.parseExpression()
			return err
		}
	} else {
		p.cursor-- // 回退游标
		_, err := p.parseExpression()
		return err
	}
}

func (p *Parser) parseAssignment() error {
	ident, err := p.parseIdentifier()
	if err != nil {
		return err
	}
	if _, ok := p.reservedWords[ident]; ok {
		return fmt.Errorf("'%s' is a reserved word", ident)
	}
	assignOp, err := p.nextToken()
	if err != nil || (assignOp.Value != ":=" && assignOp.Value != ":") {
		return fmt.Errorf("invalid assignment operator")
	}
	expr, err := p.parseExpression()
	if err != nil {
		return err
	}
	res, err := p.eval(expr)
	if err != nil {
		return err
	}
	p.symbolTable[ident] = res
	return nil
}

func (p *Parser) parseIdentifier() (string, error) {
	token, err := p.nextToken()
	if err != nil {
		return "", err
	}
	if token.Type != IDENTIFIER {
		return "", fmt.Errorf("expected identifier")
	}
	return token.Value, nil
}

func (p *Parser) parseExpression() (*Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		op, err := p.consumeOperator()
		if err != nil || (op != "+" && op != "-") {
			p.cursor-- //回退游标
			return left, nil
		}
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &Node{Type: EXPRESSION_NODE, Children: []*Node{left, right, {Type: OPERATOR_NODE, Value: op}}}
	}
}

func (p *Parser) parseTerm() (*Node, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for {
		op, err := p.consumeOperator()
		if err != nil || op != "*" && op != "/" {
			p.cursor--
			return left, nil
		}
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &Node{Type: EXPRESSION_NODE, Children: []*Node{left, right, {Type: OPERATOR_NODE, Value: op}}}
	}
}

func (p *Parser) parseFactor() (*Node, error) {
	token, err := p.nextToken()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case NUMBER:
		return &Node{Type: NUMBER_NODE, Value: token.Value}, nil
	case LPAREN:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		closingParen, err := p.nextToken()
		if err != nil || closingParen.Type != RPAREN {
			return nil, fmt.Errorf("expected ')'")
		}
		return expr, nil
	case IDENTIFIER:
		next, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if next.Type == LPAREN {
			p.cursor -= 1
			return p.parseFunctionCall(token.Value)
		} else {
			p.cursor -= 1
			if _, ok := p.reservedWords[token.Value]; ok {
				return &Node{Type: VARIABLE_NODE, Value: token.Value}, nil
			} else if val, ok := p.symbolTable[token.Value]; ok {
				return &Node{Type: SYMBOL_NODE, Value: token.Value, Result: val}, nil
			} else {
				return nil, fmt.Errorf("undefined variable or function: %s", token.Value)
			}
		}
	default:
		return nil, fmt.Errorf("unexpected token: %s", token.Value)
	}
}

func (p *Parser) parseFunctionCall(functionName string) (*Node, error) {
	node := &Node{Type: FUNCTION_NODE, Value: functionName, Children: []*Node{}}
	lparen, err := p.nextToken()
	if err != nil || lparen.Type != LPAREN {
		return nil, fmt.Errorf("expected '('")
	}

	for {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, arg)
		next, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if next.Type == RPAREN {
			break
		} else if next.Value != "," {
			return nil, fmt.Errorf("expected ',' or ')'")
		}
	}
	return node, nil
}

func (p *Parser) eval(node *Node) ([]float64, error) {
	switch node.Type {
	case NUMBER_NODE:
		num, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return nil, err
		}
		// 获取时间序列长度
		dataLen := 0
		if len(p.data) > 0 {
			for _, v := range p.data {
				dataLen = len(v)
				break // 假设所有时间序列长度相同
			}
		}
		// 扩展单个数值到时间序列长度
		res := make([]float64, dataLen)
		for i := range res {
			res[i] = num
		}
		return res, nil
	case OPERATOR_NODE:
		leftRes, err := p.eval(node.Children[0])
		if err != nil {
			return nil, err
		}
		rightRes, err := p.eval(node.Children[1])
		if err != nil {
			return nil, err
		}
		return p.applyOperator(node.Children[2].Value, leftRes, rightRes), nil
	case VARIABLE_NODE:
		if val, ok := p.data[node.Value]; ok {
			return val, nil
		} else {
			return nil, fmt.Errorf("undefined variable: %s", node.Value)
		}
	case SYMBOL_NODE:
		if val, ok := p.symbolTable[node.Value]; ok {
			return val, nil
		} else {
			return nil, fmt.Errorf("undefined symbol: %s", node.Value)
		}
	case EXPRESSION_NODE:
		leftRes, err := p.eval(node.Children[0])
		if err != nil {
			return nil, err
		}
		rightRes, err := p.eval(node.Children[1])
		if err != nil {
			return nil, err
		}
		return p.applyOperator(node.Children[2].Value, leftRes, rightRes), nil
	case FUNCTION_NODE:
		switch node.Value {
		case "MA":
			return p.evalMA(node.Children)
		case "REF":
			return p.evalREF(node.Children)
		case "HHV":
			return p.evalHHV(node.Children)
		case "LLV":
			return p.evalLLV(node.Children)
		default:
			return nil, fmt.Errorf("undefined function: %s", node.Value)
		}

	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func (p *Parser) evalREF(args []*Node) ([]float64, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("REF 函数需要两个参数")
	}
	seriesData, err := p.eval(args[0])
	if err != nil {
		return nil, err
	}
	offset, err := strconv.Atoi(args[1].Value)
	if err != nil {
		return nil, fmt.Errorf("REF 函数的第二个参数必须是整数")
	}

	res := make([]float64, len(seriesData))
	for i := range res {
		if i >= offset && i-offset < len(seriesData) && i-offset >= 0 {
			res[i] = seriesData[i-offset]
		} else {
			res[i] = math.NaN()
		}
	}
	return res, nil
}

func (p *Parser) evalMA(args []*Node) ([]float64, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("MA 函数需要两个参数")
	}
	seriesData, err := p.eval(args[0])
	if err != nil {
		return nil, err
	}
	period, err := strconv.Atoi(args[1].Value)
	if err != nil {
		return nil, fmt.Errorf("MA 函数的第二个参数必须是整数")
	}

	res := make([]float64, len(seriesData))
	for i := range res {
		sum := 0.0
		count := 0
		for j := i - period + 1; j <= i; j++ {
			if j >= 0 && j < len(seriesData) && !math.IsNaN(seriesData[j]) {
				sum += seriesData[j]
				count++
			}
		}
		if count > 0 {
			res[i] = sum / float64(count)
		} else {
			res[i] = math.NaN()
		}
	}
	return res, nil
}

func (p *Parser) evalHHV(args []*Node) ([]float64, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("HHV 函数需要两个参数")
	}
	seriesData, err := p.eval(args[0])
	if err != nil {
		return nil, err
	}
	period, err := strconv.Atoi(args[1].Value)
	if err != nil || period <= 0 {
		return nil, fmt.Errorf("HHV 函数的第二个参数必须是正整数")
	}

	res := make([]float64, len(seriesData))
	for i := range res {
		max := math.NaN()
		for j := i - period + 1; j <= i; j++ {
			if j >= 0 && j < len(seriesData) && (!math.IsNaN(seriesData[j]) && (math.IsNaN(max) || seriesData[j] > max)) {
				max = seriesData[j]
			}
		}
		res[i] = max
	}
	return res, nil
}

func (p *Parser) evalLLV(args []*Node) ([]float64, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("LLV 函数需要两个参数")
	}
	seriesData, err := p.eval(args[0])
	if err != nil {
		return nil, err
	}
	period, err := strconv.Atoi(args[1].Value)
	if err != nil || period <= 0 {
		return nil, fmt.Errorf("LLV 函数的第二个参数必须是正整数")
	}

	res := make([]float64, len(seriesData))
	for i := range res {
		min := math.NaN()
		for j := i - period + 1; j <= i; j++ {
			if j >= 0 && j < len(seriesData) && (!math.IsNaN(seriesData[j]) && (math.IsNaN(min) || seriesData[j] < min)) {
				min = seriesData[j]
			}
		}
		res[i] = min
	}
	return res, nil
}

func (p *Parser) applyOperator(op string, left, right []float64) []float64 {
	if len(left) != len(right) {
		panic("时间序列长度不匹配")
	}
	res := make([]float64, len(left))
	for i := range left {
		switch op {
		case "+":
			res[i] = left[i] + right[i]
		case "-":
			res[i] = left[i] - right[i]
		case "*":
			res[i] = left[i] * right[i]
		case "/":
			if right[i] == 0 {
				panic("除数为零")
			}
			res[i] = left[i] / right[i]
		default:
			panic(fmt.Sprintf("不支持的运算符: %s", op))
		}
	}
	return res
}

func (p *Parser) nextToken() (*Token, error) {
	if p.cursor >= len(p.tokens) {
		return nil, fmt.Errorf("no more tokens")
	}
	token := p.tokens[p.cursor]
	p.cursor++
	return &token, nil
}

func (p *Parser) consumeOperator() (string, error) {
	token, err := p.nextToken()
	if err != nil || token.Type != OPERATOR {
		return "", err
	}
	return token.Value, nil
}

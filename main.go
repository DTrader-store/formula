package main

import (
	"fmt"

	"codeup.aliyun.com/stocker/StockFormula/formula"
)

func main() {
	expression := `
	V1:=(1+CLOSE)*2;
	`
	data := map[string][]float64{
		"CLOSE": {10, 12, 15, 14, 16, 18, 20, 19, 22, 25},
	}
	lexer := formula.NewLexer(expression)
	tokens, err := lexer.Tokenize()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	parser := formula.NewParser(tokens, data)
	err = parser.ParseApp()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(parser.Result())
	}
}

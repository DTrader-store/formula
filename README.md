# 股票公式的解析器

## 1. 为什么要写这玩意
主要配合我自己量化交易使用，目前使用golang来做量化交易语言。很多交易信号都是通过公式来计算的，如果用golang硬编码这些公式，会非常不方便，
所以直接写一个解释器。

## 2. 公式的语法
支持同花顺/通达信的公式语法，但是不支持所有的函数，只支持一部分函数，后续会慢慢增加。

## 3. 目前支持的范围
- [x] 基本的数学运算
- [ ] 逻辑运算
- [ ] 函数
  - [x] HHV
  - [x] LLV
  - [x] MA
  - [x] REF
  - [ ] TODO...
- [X] 变量
- [ ] TODO...

## 使用说明
```go

func main() {
	expression := `
	V1:=(1+CLOSE)*2;
	V2:=HHV(CLOSE, 5);
	V3:=LLV(CLOSE, 5);
	V4:=MA(V1+V2+V3, 5);
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
```
```shell
go run main.go
map[
  V1:[22 26 32 30 34 38 42 40 46 52]
  V2:[10 12 15 15 16 18 20 20 22 25]
  V3:[10 10 10 10 10 12 14 14 16 18]
  V4:[42 45 49 50.5 52.4 57.6 63.2 66.6 72.4 79.4]
]
```

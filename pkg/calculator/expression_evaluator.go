package calculator

import (
	"errors"
	"math"
)

var (
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNonExistingOperation = errors.New("operation does not exist or not implemented")
	ErrCovertExample        = errors.New("line is not a mathematical expression or contains an error")
)

type Node struct {
	Num1 float64
	Num2 float64
	Sign string
}

// type Node struct {
// 	Num1     string `json:"num1"`
// 	Num2     string `json:"num2"`
// 	Sign     string `json:"sign"`
// 	Variable string `json:"variable"`
// }

func NewNode(num1, num2 float64, sign string) *Node {
	return &Node{Num1: num1, Num2: num2, Sign: sign}
}

func (n *Node) Calculate() (float64, error) {
	switch n.Sign {
	case "+":
		return n.Num1 + n.Num2, nil
	case "-":
		return n.Num1 - n.Num2, nil
	case "*":
		return n.Num1 * n.Num2, nil
	case "/":
		if n.Num2 == 0 {
			return 0, ErrDivisionByZero
		}
		return n.Num1 / n.Num2, nil
	case "^":
		return math.Pow(n.Num1, n.Num2), nil
	default:
		return 0, ErrNonExistingOperation
	}
}

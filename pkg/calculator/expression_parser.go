package calculator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/tainj/distributed_calculator2/internal/models"
)

var (
	OperatorPriority = map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
		"^": 3,
		"~": 4,
		"(": 6,
	}

	// Right-associative operators
	RightAssociative = map[string]bool{
		"^": true, // 2^3^4 = 2^(3^4)
	}
)

func NewExample(num1, num2, sign string) (models.Task, string) {
	variable := uuid.New().String() // generate variable name where result will be stored
	return models.Task{Num1: num1, Num2: num2, Sign: sign, Variable: variable}, variable
}

// Stack implementation and its methods
type Stack struct {
	list []string
}

func NewStack() *Stack {
	return &Stack{list: make([]string, 0)}
}

func (s *Stack) Push(item string) {
	s.list = append(s.list, item)
}

func (s *Stack) IsEmptyStack() bool {
	return len(s.list) == 0
}

func (s *Stack) Pop() string {
	index := len(s.list) - 1
	result := s.list[index]
	s.list = s.list[:index]
	return result
}

func (s *Stack) Peek() string {
	index := len(s.list) - 1
	return s.list[index]
}

type Expression struct {
	Infix   string // Infix expression
	Postfix string // Postfix expression
}

func NewExpression(str string) *Expression {
	return &Expression{Infix: str}
}

// Check validates expression without using govaluate.
func (s *Expression) Check() bool {
	return s.IsValidMathExpression()
}

// IsValidMathExpression checks if a string is a valid mathematical expression.
// Supports: digits, +, -, *, /, ^, ~ (unary minus), ., ()
func (s *Expression) IsValidMathExpression() bool {
	expr := strings.ReplaceAll(s.Infix, " ", "")
	if expr == "" {
		return false
	}

	// Allowed characters
	allowed := "+-*/().~^"
	for i, ch := range expr {
		if !unicode.IsDigit(ch) && !strings.ContainsRune(allowed, ch) {
			return false
		}

		// Forbid decimal point at the end
		if ch == '.' {
			if i == len(expr)-1 {
				return false
			}
			next := rune(expr[i+1])
			if !unicode.IsDigit(next) {
				return false
			}
		}
	}

	// Check brackets
	balance := 0
	for _, ch := range expr {
		if ch == '(' {
			balance++
		} else if ch == ')' {
			balance--
			if balance < 0 {
				return false // Extra closing bracket
			}
		}
	}
	if balance != 0 {
		return false // Unbalanced brackets
	}

	// Check for consecutive operators
	binaryOps := "+-*/^"
	for i := 0; i < len(expr)-1; i++ {
		curr := rune(expr[i])
		next := rune(expr[i+1])

		// Two binary operators in a row - error
		if strings.ContainsRune(binaryOps, curr) && strings.ContainsRune(binaryOps, next) {
			return false
		}

		// ~ - unary minus
		if curr == '~' {
			if i == len(expr)-1 {
				return false // ~ at the end
			}
			if next != '(' && !unicode.IsDigit(next) && next != '~' {
				return false // After ~ must be number, ( or ~
			}
		}
	}

	// Check first character
	first := rune(expr[0])
	if !isDigitOrUnaryMinusOrOpenParen(first) {
		return false
	}

	// Check last character
	last := rune(expr[len(expr)-1])
	if strings.ContainsRune("+-*/^~", last) {
		return false
	}

	return true
}

// isDigitOrUnaryMinusOrOpenParen checks if we can start an expression with this character.
func isDigitOrUnaryMinusOrOpenParen(ch rune) bool {
	return unicode.IsDigit(ch) || ch == '~' || ch == '('
}

func (s *Expression) Convert() (bool, error) {
	if !s.Check() {
		return false, fmt.Errorf("line is not a mathematical expression or contains an error")
	}
	// Initialize list, stack, and list for numbers
	list := make([]string, 0)
	stack := NewStack()
	example := strings.ReplaceAll(s.Infix, " ", "") // remove spaces
	number := make([]rune, 0)
	for _, i := range example {
		sign := string(i)
		if unicode.IsDigit(i) { // check if character is a digit
			number = append(number, i) // add to numbers list
			continue
		} else if sign == "." {
			number = append(number, rune(sign[0]))
		} else {
			if len(number) != 0 {
				list = append(list, string(number)) // if not a digit, add the whole string to the list
				number = make([]rune, 0)
			}
		}
		if value, isOperator := OperatorPriority[sign]; isOperator {
			// Process operators: +, -, *, /, ^, ~, (
			// Pop operators from stack with higher or equal priority
			// BUT: if operator is right-associative (e.g., ^), don't pop at equal priority
			for !stack.IsEmptyStack() {
				top := stack.Peek()
				if top == "(" {
					break
				}

				topPriority := OperatorPriority[top]

				if topPriority > value {
					list = append(list, stack.Pop())
				} else if topPriority == value {
					// If priority is equal
					// Check associativity: if left-associative - pop, if right - don't
					if !RightAssociative[sign] {
						list = append(list, stack.Pop())
					} else {
						break
					}
				} else {
					break
				}
			}
			stack.Push(sign) // add current operator to stack
		}
		if i == ')' { // extract operators from stack
			for stack.Peek() != "(" {
				list = append(list, stack.Pop())
			}
			stack.Pop() // remove "("
		}
	}
	if len(number) > 0 {
		list = append(list, string(number)) // add last number if present
	}
	for !stack.IsEmptyStack() {
		list = append(list, stack.Pop()) // unload remaining stack
	}
	s.Postfix = strings.Join(list, " ")
	return true, nil
}

func (s *Expression) Calculate() ([]*models.Task, string) {
	results := make([]*models.Task, 0)
	expression := strings.Split(s.Postfix, " ") // form a list of numbers and operators
	if len(expression) == 1 {
		num := expression[0]
		// consider this as: 0 + num
		result, variable := NewExample("0", num, "+")
		return append(results, &result), variable
	}
	for len(expression) != 1 {
		for index, sign := range expression {
			if _, isOperator := OperatorPriority[sign]; isOperator {
				var num1, num2 string
				var newExpr []string
				if sign == "~" {
					// unary minus: ~X → 0 - X
					if index < 1 {
						return nil, "" // error: no operand
					}
					num1 = "0"
					num2 = expression[index-1]
					sign := "-" // always subtraction
					result, variable := NewExample(num1, num2, sign)
					results = append(results, &result)
					newExpr = replaceUnary(expression, index, variable)
				} else {
					// Binary operator: +, -, *, /, ^
					if index < 2 {
						return nil, "" // error: not enough operands
					}
					num1 = expression[index-2]
					num2 = expression[index-1]
					result, variable := NewExample(num1, num2, sign)
					results = append(results, &result)
					newExpr = replaceBinary(expression, index, variable)
				}

				expression = newExpr
				break
			}
		}
	}
	return results, expression[0]
}

func replaceUnary(expr []string, opIndex int, varName string) []string {
	start := opIndex - 1
	end := opIndex + 1
	if start < 0 {
		start = 0
	}
	if end > len(expr) {
		end = len(expr)
	}
	return append(append(append([]string{}, expr[:start]...), varName), expr[end:]...)
}

func replaceBinary(expr []string, opIndex int, varName string) []string {
	start := opIndex - 2
	end := opIndex + 1
	if start < 0 {
		start = 0
	}
	if end > len(expr) {
		end = len(expr)
	}
	return append(append(append([]string{}, expr[:start]...), varName), expr[end:]...)
}

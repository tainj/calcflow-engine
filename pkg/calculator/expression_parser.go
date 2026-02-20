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

		// запрещаем точку в конце
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

		// два бинарных оператора подряд — ошибка
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

// isDigitOrUnaryMinusOrOpenParen проверяет, можно ли начать выражение с этого символа.
func isDigitOrUnaryMinusOrOpenParen(ch rune) bool {
	return unicode.IsDigit(ch) || ch == '~' || ch == '('
}

func (s *Expression) Convert() (bool, error) {
	if !s.Check() {
		return false, fmt.Errorf("line is not a mathematical expression or contains an error")
	}
	// инициализация списка, стека и списка для чисел
	list := make([]string, 0)
	stack := NewStack()
	example := strings.ReplaceAll(s.Infix, " ", "") // удаляем пробелы
	number := make([]rune, 0)
	for _, i := range example {
		sign := string(i)
		if unicode.IsDigit(i) { // проверяем является ли символ цифрой
			number = append(number, i) // добавляем в список для чисел
			continue
		} else if sign == "." {
			number = append(number, rune(sign[0]))
		} else {
			if len(number) != 0 {
				list = append(list, string(number)) // если это не цифра, то добавляем всю строку в список
				number = make([]rune, 0)
			}
		}
		if value, isOperator := OperatorPriority[sign]; isOperator {
			// обрабатываем операторы: +, -, *, /, ^, ~, (
			// выталкиваем из стека операторы с большим или равным приоритетом
			// НО: если оператор правоассоциативный (например, ^), то при равном приоритете — НЕ выталкиваем
			for !stack.IsEmptyStack() {
				top := stack.Peek()
				if top == "(" {
					break
				}

				topPriority := OperatorPriority[top]

				if topPriority > value {
					list = append(list, stack.Pop())
				} else if topPriority == value {
					// если приоритет равен
					// смотрим ассоциативность: если левоассоциативный — выталкиваем, если право — нет
					if !RightAssociative[sign] {
						list = append(list, stack.Pop())
					} else {
						break
					}
				} else {
					break
				}
			}
			stack.Push(sign) // добавляем текущий оператор в стек
		}
		if i == ')' { // извлекаем операторы из стека
			for stack.Peek() != "(" {
				list = append(list, stack.Pop())
			}
			stack.Pop() // удаляем "("
		}
	}
	if len(number) > 0 {
		list = append(list, string(number)) // добавляем последние число, если имеется
	}
	for !stack.IsEmptyStack() {
		list = append(list, stack.Pop()) // выгружаем остаток стека
	}
	s.Postfix = strings.Join(list, " ")
	return true, nil
}

func (s *Expression) Calculate() ([]*models.Task, string) {
	results := make([]*models.Task, 0)
	expression := strings.Split(s.Postfix, " ") // формируем список из чисел и операторов
	if len(expression) == 1 {
		num := expression[0]
		// считаем, что это: 0 + num
		result, variable := NewExample("0", num, "+")
		return append(results, &result), variable
	}
	for len(expression) != 1 {
		for index, sign := range expression {
			if _, isOperator := OperatorPriority[sign]; isOperator {
				var num1, num2 string
				var newExpr []string
				if sign == "~" {
					// унарный минус: ~X → 0 - X
					if index < 1 {
						return nil, "" // ошибка: нет операнда
					}
					num1 = "0"
					num2 = expression[index-1]
					sign := "-" // всегда вычитание
					result, variable := NewExample(num1, num2, sign)
					results = append(results, &result)
					newExpr = replaceUnary(expression, index, variable)
				} else {
					// Бинарный оператор: +, -, *, /, ^
					if index < 2 {
						return nil, "" // ошибка: мало операндов
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

package parser

import (
	"distributed-calculator/orchestrator/internal/storage"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidExpression  = errors.New("invalid expression")
	ErrUnsupportedOperand = errors.New("unsupported operand")
)

type Parser struct {
	operators map[string]int
}

func NewParser() *Parser {
	return &Parser{
		operators: map[string]int{
			"+": 1,
			"-": 1,
			"*": 2,
			"/": 2,
		},
	}
}

func (p *Parser) Parse(expr string) ([]*storage.Task, error) {
	tokens, err := p.tokenize(expr)
	if err != nil {
		return nil, err
	}

	postfix, err := p.infixToPostfix(tokens)
	if err != nil {
		return nil, err
	}

	return p.generateTasks(postfix)
}

func (p *Parser) tokenize(expr string) ([]string, error) {
	var tokens []string
	var numberBuffer strings.Builder

	for _, char := range strings.TrimSpace(expr) {
		switch {
		case char >= '0' && char <= '9' || char == '.':
			numberBuffer.WriteRune(char)
		case char == '(' || char == ')':
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}
			tokens = append(tokens, string(char))
		case p.isOperator(string(char)):
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}
			tokens = append(tokens, string(char))
		case char == ' ':
			continue
		default:
			return nil, fmt.Errorf("%w: invalid character '%c'", ErrInvalidExpression, char)
		}
	}

	if numberBuffer.Len() > 0 {
		tokens = append(tokens, numberBuffer.String())
	}

	return tokens, nil
}

func (p *Parser) infixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var stack []string

	for _, token := range tokens {
		switch {
		case p.isNumber(token):
			output = append(output, token)
		case token == "(":
			stack = append(stack, token)
		case token == ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, ErrInvalidExpression
			}
			stack = stack[:len(stack)-1]
		case p.isOperator(token):
			for len(stack) > 0 && p.isOperator(stack[len(stack)-1]) && 
				p.operators[stack[len(stack)-1]] >= p.operators[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedOperand, token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrInvalidExpression
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func (p *Parser) generateTasks(postfix []string) ([]*storage.Task, error) {
	var stack []string
	tasks := make([]*storage.Task, 0)
	id := 1

	for _, token := range postfix {
		if p.isNumber(token) {
			stack = append(stack, token)
			continue
		}

		if len(stack) < 2 {
			return nil, ErrInvalidExpression
		}

		arg2 := stack[len(stack)-1]
		arg1 := stack[len(stack)-2]
		stack = stack[:len(stack)-2]

		task := &storage.Task{
			ID:        fmt.Sprintf("%d", id),
			Arg1:      arg1,
			Arg2:      arg2,
			Operation: token,
			Status:    "pending",
		}

		tasks = append(tasks, task)
		stack = append(stack, task.ID)
		id++
	}

	if len(stack) != 1 {
		return nil, ErrInvalidExpression
	}

	return tasks, nil
}

func (p *Parser) isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func (p *Parser) isOperator(s string) bool {
	_, exists := p.operators[s]
	return exists
}

package rambo_test

import (
	"fmt"
	"strings"
	"testing"
)

type Calc struct{ Total float64 }

type Add struct{ Amount float64 }
type Sub struct{ Amount float64 }
type Mul struct{ Amount float64 }
type Div struct{ Amount float64 }
type Total struct{}

func (c *Add) ExecuteOn(calc *Calc) error { calc.Total += c.Amount; return nil }
func (c *Sub) ExecuteOn(calc *Calc) error { calc.Total -= c.Amount; return nil }
func (c *Mul) ExecuteOn(calc *Calc) error { calc.Total *= c.Amount; return nil }
func (c *Div) ExecuteOn(calc *Calc) error { calc.Total /= c.Amount; return nil }
func (q *Total) QueryOn(calc *Calc) any   { return calc.Total }

// ---

func assertEq[Type comparable](t *testing.T, actual Type, expected Type) {
	t.Helper()
	assert(t, actual == expected, func() string {
		return strings.Join([]string{
			red("Assertion failed: expected values to be eq (==)."), "",
			bold(blue("Actual:")), fmt.Sprintf("%+v", actual), "",
			bold(blue("Expected:")), fmt.Sprintf("%+v", expected), "",
		}, "\n")
	})
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	assert(t, err == nil, func() string {
		return strings.Join([]string{
			red("Assertion failed: expected no error."), "",
			bold(blue("Error: ")), err.Error(), "",
		}, "\n")
	})
}

func assert(t *testing.T, truth bool, lazyMessage func() string) {
	t.Helper()
	if !truth {
		message := lazyMessage()
		if len(message) == 0 {
			message = "Assertion failed!"
		}

		t.Error(message)
		t.FailNow()
	}
}

func red(text string) string   { return reset("\033[31m" + text) }
func blue(text string) string  { return reset("\033[34m" + text) }
func bold(text string) string  { return reset("\033[1m" + text) }
func reset(text string) string { return text + "\033[0m" }

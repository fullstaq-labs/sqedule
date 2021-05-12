package mocking

import (
	"fmt"
	"strings"
)

// IPrinter provides a swappable message printing interface so that during
// testing messages can be captured.
type IPrinter interface {
	Println(a ...interface{})
	Printf(format string, a ...interface{})
}

// RealPrinter prints all messages to stdout.
type RealPrinter struct{}

func (_ RealPrinter) Println(a ...interface{}) {
	fmt.Println(a...)
}

func (_ RealPrinter) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// FakePrinter prints all messages to an internal buffer. The buffer
// can be obtained by calling String().
type FakePrinter struct {
	Builder strings.Builder
}

func (p *FakePrinter) Println(a ...interface{}) {
	p.Builder.WriteString(fmt.Sprintln(a...))
}

func (p *FakePrinter) Printf(format string, a ...interface{}) {
	p.Builder.WriteString(fmt.Sprintf(format, a...))
}

func (p FakePrinter) String() string {
	return p.Builder.String()
}

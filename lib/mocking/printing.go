package mocking

import (
	"fmt"
	"os"
	"strings"
)

// IPrinter provides a swappable output/message printing interface so that during
// testing output/messages can be captured.
type IPrinter interface {
	PrintMessageln(a ...interface{})
	PrintMessagef(format string, a ...interface{})
	PrintOutputln(a ...interface{})
}

// RealPrinter prints everything to stdout.
type RealPrinter struct{}

func (_ RealPrinter) PrintMessageln(a ...interface{}) {
	fmt.Println(a...)
}

func (_ RealPrinter) PrintMessagef(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func (_ RealPrinter) PrintOutputln(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

// FakePrinter prints everything to an internal buffer. The buffer
// can be obtained by calling String().
type FakePrinter struct {
	Builder strings.Builder
}

func (p *FakePrinter) PrintMessageln(a ...interface{}) {
	p.Builder.WriteString(fmt.Sprintln(a...))
}

func (p *FakePrinter) PrintMessagef(format string, a ...interface{}) {
	p.Builder.WriteString(fmt.Sprintf(format, a...))
}

func (p *FakePrinter) PrintOutputln(a ...interface{}) {
	p.Builder.WriteString(fmt.Sprintln(a...))
}

func (p FakePrinter) String() string {
	return p.Builder.String()
}

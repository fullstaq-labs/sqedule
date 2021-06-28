package cli

import "github.com/fullstaq-labs/sqedule/lib/mocking"

func PrintSeparatorln(printer mocking.IPrinter) {
	printer.PrintMessageln("--------------------")
}

func PrintCelebrationlnf(printer mocking.IPrinter, format string, a ...interface{}) {
	printer.PrintMessagef("🎉 "+format+"\n", a...)
}

func PrintTiplnf(printer mocking.IPrinter, format string, a ...interface{}) {
	printer.PrintMessagef("💡 "+format+"\n", a...)
}

func PrintCaveatlnf(printer mocking.IPrinter, format string, a ...interface{}) {
	printer.PrintMessagef("⚠️  "+format+"\n", a...)
}

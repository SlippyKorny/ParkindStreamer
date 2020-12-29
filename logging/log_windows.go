package logging

import (
	"fmt"
	"time"
)

func log(a []string) {
	t := time.Now()
	fmt.Printf("[%s]:", t.Format("02/01/2006 15:04:05"))
	for _, foo := range a {
		fmt.Print(" ", foo)
	}
	fmt.Println()
}

func NormalLog(a ...string) {
	log(a)
}

func InfoLog(a ...string) {
	if !args.verbosity {
		return
	}

	log(a)
}

func SuccessLog(a ...string) {
	log(a)
}

func WarningLog(a ...string) {
	log(a)
}

func ErrorLog(a ...string) {
	log(a)
}

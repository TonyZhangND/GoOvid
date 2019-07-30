package commons

import (
	"fmt"
	"os"
	"runtime/debug"
)

type ProcessID uint16
type PortNum uint16

// fatalOvidError prints the error messange and kills the entire program
func FatalOvidError(errMsg string) {
	fmt.Printf("Error : Ovid : %v\n", errMsg)
	debug.PrintStack()
	os.Exit(1)
}

func CheckFatalOvidError(err error, errMsg string) {
	if err != nil {
		fmt.Printf("Error : Ovid : %v\n", errMsg)
		debug.PrintStack()
		os.Exit(1)
	}
}

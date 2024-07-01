package main

import (
	"os"

	"github.com/ravilock/scargo/cmd"
	"github.com/ravilock/scargo/cmd/exit"
)

func recoverPanicError() {
	if r := recover(); r != nil {
		if e, ok := r.(*exit.PanicError); ok {
			os.Exit(e.Code)
		}
		panic(r)
	}
}

func main() {
	defer recoverPanicError()
	cmd.Execute()
}

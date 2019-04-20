package main

import (
	"fmt"
	"os"

	mimeapps "github.com/stephane-martin/go-mimeapps"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	err = mimeapps.OpenRemote(filename, f)
	_ = f.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

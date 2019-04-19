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
	app, err := mimeapps.FilenameToApplication(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(app)
}

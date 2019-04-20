package mimeapps

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gdamore/tcell"
)

func isPlaceHolder(s string) bool {
	return s == `%f` || s == `%F` || s == `%u` || s == `%U`
}

func OpenLocal(filename string) error {
	opener, terminal, err := FilenameToApplication(filename)
	if err != nil {
		return err
	}
	if opener == nil {
		return fmt.Errorf("no opener found for: %s", filename)
	}
	stats, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if stats.IsDir() {
		return fmt.Errorf("is a directory: %s", filename)
	}
	var found bool
	for i := range opener {
		if isPlaceHolder(opener[i]) {
			opener[i] = filename
			found = true
		}
	}
	if !found {
		opener = append(opener, filename)
	}
	cmd := exec.Command(opener[0], opener[1:]...)
	if terminal {
		scr, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		err = scr.Init()
		if err != nil {
			return err
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		scr.Fini()
		return err
	}

	return cmd.Run()
}

func OpenRemote(filename string, r io.Reader) error {
	tempDir, err := ioutil.TempDir("", "mimeapps-opener")
	if err != nil {
		return err
	}
	baseName := filepath.Base(filename)
	destName := filepath.Join(tempDir, baseName)
	dest, err := os.Create(destName)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return err
	}
	_, err = io.Copy(dest, r)
	if err != nil {
		_ = dest.Close()
		_ = os.RemoveAll(tempDir)
		return err
	}
	_ = dest.Close()
	_ = os.Chmod(destName, 0500)
	return OpenLocal(destName)
}

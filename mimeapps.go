package mimeapps

import (
	"errors"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
)

func MimeTypeToDesktop(mimetype string) (string, error) {
	var section *ini.Section
	list, err := MimeAppsPathsList()
	if err != nil {
		return "", err
	}
	for _, path := range list {
		f, err := ini.Load(path)
		if err != nil {
			return "", err
		}
		section, err = f.GetSection("Default Applications")
		if err == nil {
			k := section.Key(mimetype)
			if k != nil && k.String() != "" {
				return k.String(), nil
			}
		}
	}
	list, err = DefaultsPathsList()
	if err != nil {
		return "", err
	}
	for _, path := range list {
		f, err := ini.Load(path)
		if err != nil {
			return "", err
		}
		if filepath.Base(path) == "defaults.list" {
			section, err = f.GetSection("Default Applications")
		} else {
			section, err = f.GetSection("MIME Cache")
		}
		if err == nil {
			k := section.Key(mimetype)
			if k != nil && k.String() != "" {
				return strings.SplitN(k.String(), ";", 2)[0], nil
			}
		}

	}

	return "", nil
}

var ErrFound = errors.New("found")

func MimeTypeToDesktopPath(mimetype string) (string, error) {
	xdg, err := XDG()
	if err != nil {
		return "", err
	}
	desktopFile, err := MimeTypeToDesktop(mimetype)
	if err != nil || desktopFile == "" {
		return "", err
	}
	var search []string
	d := filepath.Join(xdg.DataHome, "applications")
	if DirExists(d) {
		search = append(search, d)
	}
	for _, d := range xdg.DataDirs {
		d = filepath.Join(d, "applications")
		if DirExists(d) {
			search = append(search, d)
		}
	}
	desktopFilePath := ""
	for _, dirname := range search {
		err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if filepath.Base(path) == desktopFile {
				desktopFilePath = path
				return ErrFound
			}
			return nil
		})
		if err == ErrFound {
			break
		}
	}
	return desktopFilePath, nil
}

func MimeTypeToApplication(mimetype string) (string, bool, error) {
	path, err := MimeTypeToDesktopPath(mimetype)
	if err != nil || path == "" {
		return path, false, err
	}
	f, err := ini.Load(path)
	if err != nil {
		return "", false, err
	}
	section, err := f.GetSection("Desktop Entry")
	if err != nil {
		return "", false, nil
	}
	k := section.Key("Exec")
	if k != nil && k.String() != "" {
		kt := section.Key("Terminal")
		if kt != nil && kt.String() == "true" {
			return k.String(), true, nil
		}
		return k.String(), false, nil
	}
	return "", false, nil
}

func FilenameToApplication(filename string) ([]string, bool, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return nil, false, nil
	}
	m := mime.TypeByExtension(strings.ToLower(ext))
	if m == "" {
		return nil, false, nil
	}
	m = strings.SplitN(m, ";", 2)[0]
	app, terminal, err := MimeTypeToApplication(m)
	if err != nil {
		return nil, false, err
	}
	if app == "" {
		return nil, false, nil
	}
	return Scan(app), terminal, nil
}

func Scan(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	s = strings.Replace(s, `\\`, `\`, -1)
	var tokens []string
	var token strings.Builder
	var insideQuotes bool
	runes := []rune(s)
	nbRunes := len(runes)
	for i := 0; i < nbRunes; i++ {
		current := string(runes[i])
		next := ""
		if i < (nbRunes - 1) {
			next = string(runes[i+1])
		}

		if insideQuotes {
			if current == `\` {
				// escaped special char
				token.WriteString(next)
				i++
			} else if current == `"` {
				// end of quoted string
				t := strings.TrimSpace(token.String())
				if t != "" {
					tokens = append(tokens, t)
				}
				token.Reset()
				insideQuotes = false
			} else {
				token.WriteString(current)
			}
		} else {
			if current == `"` {
				// beginning of quoted string
				token.Reset()
				insideQuotes = true
			} else if current == ` ` {
				// next token
				t := strings.TrimSpace(token.String())
				if t != "" {
					tokens = append(tokens, t)
				}
				token.Reset()
			} else {
				token.WriteString(current)
			}

		}
	}
	t := strings.TrimSpace(token.String())
	if t != "" {
		tokens = append(tokens, t)
	}

	return tokens
}

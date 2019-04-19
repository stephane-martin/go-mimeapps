package mimeapps

import (
	"errors"
	"mime"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
)

func MimeTypeToDesktop(mimetype string) (string, error) {
	list, err := MimeAppsPathsList()
	if err != nil {
		return "", err
	}
	for _, path := range list {
		f, err := ini.Load(path)
		if err != nil {
			return "", err
		}
		section, err := f.GetSection("Default Applications")
		if err == nil {
			k := section.Key(mimetype)
			if k != nil {
				return k.String(), nil
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

func MimeTypeToApplication(mimetype string) (string, error) {
	path, err := MimeTypeToDesktopPath(mimetype)
	if err != nil || path == "" {
		return path, err
	}
	f, err := ini.Load(path)
	if err != nil {
		return "", err
	}
	section, err := f.GetSection("Desktop Entry")
	if err != nil {
		return "", nil
	}
	k := section.Key("Exec")
	if k == nil {
		return "", nil
	}
	return k.String(), nil
}

func FilenameToApplication(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", nil
	}
	m := mime.TypeByExtension(ext)
	if m == "" {
		return "", nil
	}
	return MimeTypeToApplication(m)
}

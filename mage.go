//go:build mage
// +build mage

package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	serviceDirectory       = "services"
	serviceSourceDirectory = "internal/services"
)

func All() error {
	err := Test()
	if err != nil {
		return err
	}
	return Build()
}

func Build() error {
	return exec.Command("go", "build", "cmd/moody-core").Run()
}

func Install() error {
	return exec.Command("go", "install", "cmd/moody-core").Run()
}

func Test() error {
	return exec.Command("go", "test", "./...").Run()
}

func Services() error {
	if err := os.Mkdir(serviceDirectory, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	err := filepath.WalkDir(serviceSourceDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() != serviceDirectory {
			cmd := exec.Command("go", "build", "-buildmode", "plugin")
			cmd.Dir = serviceSourceDirectory + "/" + d.Name()
			err := cmd.Run()
			if err != nil {
				return err
			}
			targetName := "/" + d.Name() + ".so"
			err = moveFile(cmd.Dir+targetName, serviceDirectory+targetName)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func moveFile(inName string, outName string) error {
	dest, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer dest.Close()

	src, err := os.Open(inName)
	if err != nil {
		return err
	}

	if _, err = io.Copy(dest, src); err != nil {
		_ = os.Remove(outName)
		return err
	}
	src.Close()

	err = os.Remove(inName)
	if err != nil {
		return err
	}
	return nil
}

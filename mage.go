//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"go/build"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	binaryName             = "moody-core"
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
	return command("go", "build", "-ldflags", "-s -w", "./cmd/moody-core").Run()
}

func Clean() error {
	return os.Remove(binaryName)
}

func Install() error {
	return command("go", "install", "-ldflags", "-s -w", "./cmd/moody-core").Run()
}

func Uninstall() error {
	path := fmt.Sprintf("%s/bin/%s", build.Default.GOPATH, binaryName)
	return os.Remove(path)
}

func Test() error {
	return command("go", "test", "./...").Run()
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
			cmd := command("go", "build", "-buildmode", "plugin")
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

func command(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
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

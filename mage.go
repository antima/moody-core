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

// Build the application into a single binary for the target architecture
func Build() error {
	return command("go", "build", "-ldflags", "-s -w", "./cmd/moody-core").Run()
}

// Buildrpi builds the application into a single binary targeting a 32-bit Linux on Raspberry Pi
func Buildrpi() error {
	cmd := command("go", "build", "-ldflags", "-s -w", "./cmd/moody-core")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm", "GOARM=7")
	return cmd.Run()
}

// Clean removes previously built binaries
func Clean() error {
	return os.Remove(binaryName)
}

// Install the application after building it, by placing it in GOPATH/bin
func Install() error {
	if err := os.MkdirAll("/etc/moody", 0755); err != nil {
		return err
	}

	if err := os.MkdirAll("/usr/local/lib/moody", 0755); err != nil {
		return err
	}

	if err := copyFile("./config/conf.json", "/etc/moody/conf.json"); err != nil {
		return err
	}

	if err := copyFile("./config/systemd/moody.service", "/etc/systemd/system/moody.service"); err != nil {
		return err
	}

	if err := moveFile(binaryName, "/usr/local/bin/"+binaryName); err != nil {
		return err
	}

	return command("systemctl enable moody").Run()
}

// Uninstall the application by removing the binary form GOPATH/bin
func Uninstall() error {
	path := fmt.Sprintf("%s/bin/%s", build.Default.GOPATH, binaryName)
	return os.Remove(path)
}

// Test the application by running go test in each sub-directory
func Test() error {
	return command("go", "test", "./...").Run()
}

// Services build every service found within ./internal/services/
func Services() error {
	if err := os.Mkdir(serviceDirectory, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	err := filepath.WalkDir(serviceSourceDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() != serviceDirectory {
			cmd := command("go", "build", "-ldflags", "-s -w", "-buildmode", "plugin")
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

func copyFile(inName string, outName string) error {
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
	return nil
}

func moveFile(inName string, outName string) error {
	if err := copyFile(inName, outName); err != nil {
		return err
	}

	if err := os.Remove(inName); err != nil {
		return err
	}
	return nil
}

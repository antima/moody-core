//go:build mage
// +build mage

package main

import "os/exec"

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

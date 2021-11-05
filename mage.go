//go:build mage
// +build mage

package main

import "os/exec"

func Build() error {
	return exec.Command("go", "build", "cmd/moody-core").Run()
}

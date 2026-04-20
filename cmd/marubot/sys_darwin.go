//go:build darwin
package main

import "syscall"

func hideConsole() {}

func getSysProcAttr() *syscall.SysProcAttr {
	return nil
}

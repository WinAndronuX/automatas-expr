package utils

import (
	"log"
	"os"
	"os/exec"
	"runtime"
)


var cls map[string]func()

func init() {

	cls = make(map[string]func())

	cls["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout= os.Stdout
		cmd.Run()
	}

	cls["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func clear() {
	f, ok := cls[runtime.GOOS]

	if !ok {
		log.Fatalf("platform %s not supported", runtime.GOOS)
	}

	f()
}

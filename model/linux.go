package model

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Linux model
type Linux struct {
	Command   string
	Args      []string
	RawOutput bool
}

//ExecuteBash returns
func (r *Linux) ExecuteBash() (string, error) {
	log.Info("Executing linux command")
	log.Info(r.Command + " " + strings.Join(r.Args, " "))
	cmd := exec.Command(r.Command, r.Args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Error(exitErr)
		} else {
			log.Error(err)
			os.Exit(1)
		}
		return "", err
	}
	log.Info("Executed linux command")
	return string(out), nil
}

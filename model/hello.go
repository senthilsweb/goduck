package model

import log "github.com/sirupsen/logrus"

// Say "Hello World"
func Say(payload string) (string, error) {
	log.Info(payload)
	return "Message received and processed", nil
}

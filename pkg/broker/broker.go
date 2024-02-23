package broker

import (
	"fmt"
	cfg "templrjs/pkg/config"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	log "github.com/sirupsen/logrus"
)

var (
	Broker *machinery.Server
)

func Setup() {

	cfg := cfg.GetConfig()

	var cnf = config.Config{
		Broker:        fmt.Sprintf("%s://%s:%s", cfg.Queue.Broker, cfg.Queue.Host, cfg.Queue.Port),
		ResultBackend: fmt.Sprintf("%s://%s:%s", cfg.Queue.Broker, cfg.Queue.Host, cfg.Queue.Port),
	}

	server, err := machinery.NewServer(&cnf)
	if err != nil {
		log.Fatal(err, "Can not create async server!")
	}

	log.Info(server.GetBroker().GetPendingTasks("templrjs_tasks"))
	Broker = server
}

// GetDB helps you to get a connection
func GetBroker() *machinery.Server {
	return Broker
}

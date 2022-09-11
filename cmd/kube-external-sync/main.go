package main

import (
	"os"

	"github.com/alehechka/kube-external-sync/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.App().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

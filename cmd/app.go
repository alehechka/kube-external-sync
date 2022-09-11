package cmd

import (
	kubeexternalsync "github.com/alehechka/kube-external-sync"
	"github.com/urfave/cli/v2"
)

// App represents the CLI application
func App() *cli.App {
	app := cli.NewApp()
	app.Version = kubeexternalsync.Version
	app.Usage = "Automatically synchronize k8s Services as ExternalName Services across namespaces."
	app.Commands = []*cli.Command{
		StartCommand,
	}

	return app
}

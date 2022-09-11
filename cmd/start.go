package cmd

import (
	"path/filepath"

	"github.com/alehechka/kube-external-sync/client"
	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"
)

const (
	debugFlag        = "debug"
	outOfClusterFlag = "out-of-cluster"
	kubeconfigFlag   = "kubeconfig"
	podNamespace     = "pod-namespace"
)

func kubeconfig() *cli.StringFlag {
	kubeconfig := &cli.StringFlag{Name: kubeconfigFlag}
	if home := homedir.HomeDir(); home != "" {
		kubeconfig.Value = filepath.Join(home, ".kube", "config")
		kubeconfig.Usage = "(optional) absolute path to the kubeconfig file"
	} else {
		kubeconfig.Usage = "absolute path to the kubeconfig file (required if running OutOfCluster)"
	}
	return kubeconfig
}

var startFlags = []cli.Flag{
	kubeconfig(),
	&cli.BoolFlag{
		Name:    debugFlag,
		Usage:   "Log debug messages.",
		EnvVars: []string{"DEBUG"},
	},
	&cli.StringFlag{
		Name:    podNamespace,
		Usage:   "Specifies the namespace that current application pod is running in.",
		EnvVars: []string{"POD_NAMESPACE"},
	},
	&cli.BoolFlag{
		Name:    outOfClusterFlag,
		Usage:   "Will use the default ~/.kube/config file on the local machine to connect to the cluster externally.",
		Aliases: []string{"local"},
	},
}

func startKubeSecretSync(ctx *cli.Context) (err error) {
	if ctx.Bool(debugFlag) {
		log.SetLevel(log.DebugLevel)
	}

	return client.SyncExternals(&client.SyncConfig{
		PodNamespace: ctx.String(podNamespace),

		OutOfCluster: ctx.Bool(outOfClusterFlag),
		KubeConfig:   ctx.String(kubeconfigFlag),
	})
}

// StartCommand starts the kube-external-sync process.
var StartCommand = &cli.Command{
	Name:   "start",
	Usage:  "Start the kube-external-sync application.",
	Action: startKubeSecretSync,
	Flags:  startFlags,
}

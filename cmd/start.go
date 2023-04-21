package cmd

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/alehechka/kube-external-sync/client"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"
)

const (
	logLevelFlag               = "log-level"
	logFormatFlag              = "log-format"
	outOfClusterFlag           = "out-of-cluster"
	kubeconfigFlag             = "kubeconfig"
	podNamespaceFlag           = "pod-namespace"
	livenessPortFlag           = "liveness-port"
	resyncPeriodFlag           = "resync-period"
	defaultIngressHostnameFlag = "default-ingress-hostname"
	enableTraefikFlag          = "enable-traefik"
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
	&cli.StringFlag{
		Name:    logLevelFlag,
		Usage:   "Log level (trace, debug, info, warn, error)",
		EnvVars: []string{"LOG_LEVEL"},
		Value:   "info",
	},
	&cli.StringFlag{
		Name:    logFormatFlag,
		Usage:   "Log format (plain, json)",
		EnvVars: []string{"LOG_FORMAT"},
		Value:   "plain",
	},
	&cli.IntFlag{
		Name:    livenessPortFlag,
		Aliases: []string{"p"},
		EnvVars: []string{"LIVENESS_PORT"},
		Usage:   "Specifies the port the listen on for the liveness probe.",
		Value:   8080,
	},
	&cli.StringFlag{
		Name:    resyncPeriodFlag,
		EnvVars: []string{"RESYNC_PERIOD"},
		Usage:   "Resynchronization period for the kubelet watcher",
		Value:   "30m",
	},
	&cli.StringFlag{
		Name:    defaultIngressHostnameFlag,
		EnvVars: []string{"DEFAULT_INGRESS_HOSTNAME"},
		Usage:   "Default hostname to use when syncing an Ingress or IngressRoute resource and proper annotation is not provided. If this value is left blank, then the hostname will be extracted from the resource being synced.",
		Value:   "30m",
	},
	&cli.BoolFlag{
		Name:    enableTraefikFlag,
		Usage:   "Enables the controller to replicate Traefik CRDs.",
		EnvVars: []string{"ENABLE_TRAEFIK"},
	},
	&cli.StringFlag{
		Name:    podNamespaceFlag,
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
	PrepareLogger(ctx)

	resyncPeriod, err := time.ParseDuration(ctx.String(resyncPeriodFlag))
	if err != nil {
		return err
	}

	return client.SyncExternals(&client.SyncConfig{
		PodNamespace: ctx.String(podNamespaceFlag),

		LivenessPort:           ctx.Int(livenessPortFlag),
		ResyncPeriod:           resyncPeriod,
		DefaultIngressHostname: ctx.String(defaultIngressHostnameFlag),
		EnableTraefik:          ctx.Bool(enableTraefikFlag),

		OutOfCluster: ctx.Bool(outOfClusterFlag),
		KubeConfig:   ctx.String(kubeconfigFlag),
	})
}

func PrepareLogger(ctx *cli.Context) {
	switch strings.ToUpper(strings.TrimSpace(ctx.String(logLevelFlag))) {
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "WARN", "WARNING":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if strings.ToUpper(strings.TrimSpace(ctx.String(logFormatFlag))) == "JSON" {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

// StartCommand starts the kube-external-sync process.
var StartCommand = &cli.Command{
	Name:   "start",
	Usage:  "Start the kube-external-sync application.",
	Action: startKubeSecretSync,
	Flags:  startFlags,
}

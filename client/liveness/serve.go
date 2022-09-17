package liveness

import (
	"fmt"
	"net/http"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
	log "github.com/sirupsen/logrus"
)

func Serve(port int, replicators []common.Replicator) error {
	h := Handler{
		Replicators: replicators,
	}

	log.Infof("starting liveness monitor on port: %d", port)

	http.Handle("/healthz", &h)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

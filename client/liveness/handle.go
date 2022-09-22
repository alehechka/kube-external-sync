package liveness

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alehechka/kube-external-sync/client/replicate/common"
)

type response struct {
	NotReady []string `json:"notReady"`
}

// Handler implements a HTTP response handler that reports on the current
// liveness status of the controller
type Handler struct {
	Replicators []common.Replicator
}

func (h *Handler) notReadyComponents() []string {
	notReady := make([]string, 0)

	for _, replicator := range h.Replicators {
		if replicator == nil {
			continue
		}

		if synced := replicator.Synced(); !synced {
			notReady = append(notReady, fmt.Sprintf("%T", replicator))
		}
	}

	return notReady
}

func (h *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r := response{
		NotReady: h.notReadyComponents(),
	}

	if len(r.NotReady) > 0 {
		res.WriteHeader(http.StatusServiceUnavailable)
	} else {
		res.WriteHeader(http.StatusOK)
	}

	enc := json.NewEncoder(res)
	_ = enc.Encode(&r)
}

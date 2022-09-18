start:
	go run cmd/kube-external-sync/main.go start \
		--local --log-level debug \
		--liveness-port 9090 \
		--pod-namespace kube-external-sync

generate:
	controller-gen object paths=./...
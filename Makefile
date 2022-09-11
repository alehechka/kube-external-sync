start:
	go run cmd/kube-external-sync/main.go start \
		--local --debug \
		--pod-namespace kube-external-sync

generate:
	controller-gen object paths=./...
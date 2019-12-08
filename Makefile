REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

default: vendor
	@echo "=============Building============="
	go build -mod vendor -o dist/$(NAME) ./cmd/predictive-horizontal-pod-autoscaler
	cp LICENSE dist/LICENSE

unittest: vendor
	@echo "=============Running unit tests============="
	go test ./... -cover -mod=vendor -coverprofile unit_cover.out --tags=unit

lint: vendor
	@echo "=============Linting============="
	go list -mod=vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

docker: default
	@echo "=============Building docker images============="
	docker build -f build/Dockerfile -t $(REGISTRY)/$(NAME):$(VERSION) .

vendor:
	go mod vendor
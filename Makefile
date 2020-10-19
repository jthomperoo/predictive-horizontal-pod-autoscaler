REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

default:
	@echo "=============Building============="
	go mod vendor
	go build -mod vendor -o dist/$(NAME) ./cmd/predictive-horizontal-pod-autoscaler
	cp LICENSE dist/LICENSE

unittest:
	@echo "=============Running unit tests============="
	go mod vendor
	go test ./... -cover -mod=vendor -coverprofile unit_cover.out --tags=unit

lint:
	@echo "=============Linting============="
	go mod vendor
	go list -mod=vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

docker: default
	@echo "=============Building docker images============="
	docker build -f build/Dockerfile -t $(REGISTRY)/$(NAME):$(VERSION) .

doc:
	@echo "=============Serving docs============="
	mkdocs serve

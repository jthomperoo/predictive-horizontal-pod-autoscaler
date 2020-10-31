REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

default:
	@echo "=============Building============="
	go mod vendor
	go build -mod vendor -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

test:
	@echo "=============Running unit tests============="
	go mod vendor
	go test ./... -cover -mod=vendor -coverprofile unit_cover.out --tags=unit
	pytest algorithms/

lint:
	@echo "=============Linting============="
	go mod vendor
	go list -mod=vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status
	pylint algorithms --rcfile=.pylintrc

docker: default
	@echo "=============Building docker images============="
	docker build -f build/Dockerfile -t $(REGISTRY)/$(NAME):$(VERSION) .

doc:
	@echo "=============Serving docs============="
	mkdocs serve

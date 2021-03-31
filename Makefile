REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

default: vendor_modules
	@echo "=============Building============="
	go build -mod vendor -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

test: vendor_modules
	@echo "=============Running unit tests============="
	go test ./... -cover -mod=vendor -coverprofile unit_cover.out --tags=unit
	pytest algorithms/

lint: vendor_modules
	@echo "=============Linting============="
	go list -mod=vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status
	pylint algorithms --rcfile=.pylintrc

beautify: vendor_modules
	@echo "=============Beautifying============="
	gofmt -s -w .

docker: default
	@echo "=============Building docker images============="
	docker build -f build/Dockerfile -t $(REGISTRY)/$(NAME):$(VERSION) .

doc:
	@echo "=============Serving docs============="
	mkdocs serve

vendor_modules:
	go mod vendor

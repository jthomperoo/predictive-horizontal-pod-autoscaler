REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

default: vendor_modules
	@echo "=============Building============="
	go build -mod vendor -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

test: vendor_modules
	@echo "=============Running unit tests============="
	go test ./... -cover -mod=vendor -coverprofile application_coverage.out --tags=unit
	pytest algorithms/ --cov-report term --cov-report=xml:algorithm_coverage.out --cov-report=html:.algorithm_coverage --cov=algorithms/

lint: vendor_modules
	@echo "=============Linting============="
	go list -mod=vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status
	pylint algorithms --rcfile=.pylintrc

beautify: vendor_modules
	@echo "=============Beautifying============="
	gofmt -s -w .
	go mod tidy
	find algorithms -name '*.py' -print0 | xargs -0 yapf -i

docker:
	@echo "=============Building docker images============="
	docker build -f build/Dockerfile -t $(REGISTRY)/$(NAME):$(VERSION) .

doc:
	@echo "=============Serving docs============="
	mkdocs serve

view_coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=application_coverage.out
	python -m webbrowser file://$(shell pwd)/.algorithm_coverage/index.html

vendor_modules:
	go mod vendor

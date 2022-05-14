REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

default:
	@echo "=============Building============="
	go build -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

test:
	@echo "=============Running tests============="
	go test ./... -cover -coverprofile application_coverage.out
	pytest algorithms/ --cov-report term --cov-report=xml:algorithm_coverage.out --cov-report=html:.algorithm_coverage --cov=algorithms/

lint:
	@echo "=============Linting============="
	staticcheck ./...
	pylint algorithms --rcfile=.pylintrc

format:
	@echo "=============Formatting============="
	gofmt -s -w .
	go mod tidy
	find algorithms -name '*.py' -print0 | xargs -0 yapf -i

docker:
	@echo "=============Building docker images============="
	docker build -f build/Dockerfile -t $(REGISTRY)/$(NAME):$(VERSION) .

doc:
	@echo "=============Serving docs============="
	mkdocs serve

coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=application_coverage.out
	python -m webbrowser file://$(shell pwd)/.algorithm_coverage/index.html

REGISTRY = jthomperoo
NAME = predictive-horizontal-pod-autoscaler
VERSION = latest

LOCAL_HELM_CHART_NAME=predictive-horizontal-pod-autoscaler-operator

run: deploy py_dependencies
	go run github.com/cosmtrek/air

py_dependencies:
	python -m pip install -r requirements-dev.txt

undeploy: generate
	helm uninstall $(LOCAL_HELM_CHART_NAME)

deploy: generate
	helm upgrade --install --set mode=development $(LOCAL_HELM_CHART_NAME) helm/

lint: generate
	@echo "=============Linting============="
	go run honnef.co/go/tools/cmd/staticcheck@v0.4.2 ./...
	pylint algorithms --rcfile=.pylintrc

format:
	@echo "=============Formatting============="
	gofmt -s -w .
	go mod tidy
	find algorithms -name '*.py' -print0 | xargs -0 yapf -i

test: gotest pytest

gotest:
	export GOCOVERDIR='.' && go test ./... -cover -coverprofile unit_cover.out

pytest:
	pytest algorithms/ --cov-report term --cov-report=xml:algorithm_coverage.out --cov-report=html:.algorithm_coverage --cov=algorithms/

docker:
	docker build . -t $(REGISTRY)/$(NAME):$(VERSION)

generate: get_controller-gen
	@echo "=============Generating Golang and YAML============="
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
	controller-gen rbac:roleName=predictive-horizontal-pod-autoscaler webhook crd:allowDangerousTypes=true \
		paths="./..." \
		output:crd:artifacts:config=helm/templates/crd \
		output:rbac:artifacts:config=helm/templates/cluster \
		output:webhook:artifacts:config=helm/templates/cluster

view_coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=unit_cover.out
	python -m webbrowser file://$(shell pwd)/.algorithm_coverage/index.html

get_controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2

doc:
	@echo "=============Serving docs============="
	mkdocs serve


# Set some variables the operator expects to have in order to work
# Those need to be the same as in deploy/ironic_ci.env
export OPERATOR_NAME=baremetal-operator
export DEPLOY_KERNEL_URL=http://172.22.0.1:6180/images/ironic-python-agent.kernel
export DEPLOY_RAMDISK_URL=http://172.22.0.1:6180/images/ironic-python-agent.initramfs
export IRONIC_ENDPOINT=http://localhost:6385/v1/
export IRONIC_INSPECTOR_ENDPOINT=http://localhost:5050/v1/
export GO111MODULE=on
export GOFLAGS=

RUN_NAMESPACE = metal3

# See version/version.go for details
GIT_COMMIT="$(shell git rev-parse --verify 'HEAD^{commit}')"
export LDFLAGS="-X github.com/metal3-io/baremetal-operator/version.Raw=$(shell git describe --always --abbrev=40 --dirty) -X github.com/metal3-io/baremetal-operator/version.Commit=${GIT_COMMIT}"

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Use our vendored copy of controller-gen
CONTROLLER_GEN=./bin/controller-gen

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet lint sec manifests unit

.PHONY:
unit:
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -ldflags $(LDFLAGS) -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt manifests
	go run -ldflags $(LDFLAGS) ./main.go -namespace $(RUN_NAMESPACE)

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

.PHONY: lint
lint: golint-binary ## Run golint
	find . -path ./tools -prune -o -type f -name \*.go \
		|grep -v zz_ \
		| xargs -L1 golint -set_exit_status

.PHONY: sec
sec: $(GOPATH)/bin/gosec
	gosec -severity medium --confidence medium -quiet -exclude-dir=tools ./...
$(GOPATH)/bin/gosec:
	go get -u github.com/securego/gosec/cmd/gosec

.PHONY: golint-binary
golint-binary:
	which golint 2>&1 >/dev/null || $(MAKE) $GOPATH/bin/golint
$GOPATH/bin/golint:
	go get -u golang.org/x/lint/golint

# Generate code
generate: $(CONTROLLER_GEN) openapi
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY:
openapi:
	openapi-gen \
		--input-dirs ./api/v1alpha1 \
		--output-package ./api/v1alpha1 \
		--output-base "" \
		--output-file-base zz_generated.openapi \
		--report-filename "-" \
		--go-header-file ./hack/boilerplate.go.txt

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

$(CONTROLLER_GEN): ./bin
	cd tools/controller-tools && go build ./cmd/controller-gen
	cp tools/controller-tools/controller-gen $@

./bin:
	mkdir -p $@

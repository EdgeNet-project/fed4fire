IMAGE := slok/kube-code-generator:v1.21.1

DIRECTORY := $(PWD)
PROJECT_PACKAGE := github.com/EdgeNet-project/fed4fire
DEPS_CMD := go mod tidy

default: generate

.PHONY: generate
generate: generate-client generate-crd

.PHONY: generate-client
generate-client:
	docker run -it --rm \
	-v $(DIRECTORY):/go/src/$(PROJECT_PACKAGE) \
	-e PROJECT_PACKAGE=$(PROJECT_PACKAGE) \
	-e CLIENT_GENERATOR_OUT=$(PROJECT_PACKAGE)/pkg/generated \
	-e APIS_ROOT=$(PROJECT_PACKAGE)/pkg/apis \
	-e GROUPS_VERSION="fed4fire:v1" \
	-e GENERATION_TARGETS="deepcopy,client" \
	$(IMAGE)

.PHONY: generate-crd
generate-crd:
	docker run -it --rm \
	-v $(DIRECTORY):/src \
	-e GO_PROJECT_ROOT=/src \
	-e CRD_TYPES_PATH=/src/pkg/apis \
	-e CRD_OUT_PATH=/src/crd \
	$(IMAGE) update-crd.sh

.PHONY: deps
deps:
	$(DEPS_CMD)

.PHONY: clean
clean:
	echo "Cleaning generated files..."
	rm -rf ./crd
	rm -rf ./pkg/generated
	rm -rf ./pkg/apis/fed4fire/v1/zz_generated.deepcopy.go

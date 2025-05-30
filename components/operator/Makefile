# local variables
PROJECT_ROOT = ../..

include ${PROJECT_ROOT}/hack/tools.mk
include ${PROJECT_ROOT}/hack/help.mk

# incoming variables

# Image URL to use all building/pushing image targets
IMG ?= europe-docker.pkg.dev/kyma-project/prod/dockerregistry-operator:main

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ Development
CONFIG_OPERATOR = $(PROJECT_ROOT)/config/operator/base
CONFIG_OPERATOR_DEV = $(PROJECT_ROOT)/config/operator/dev

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=operator-role crd webhook paths="./..." output:crd:artifacts:config=$(CONFIG_OPERATOR)/crd/bases output:rbac:artifacts:config=$(CONFIG_OPERATOR)/rbac

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="$(PROJECT_ROOT)/hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	 go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet kubebuilder-assets ## Run unit tests.
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate fmt vet ## Build operator binary.
	go build -o bin/operator main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: manifests generate ## Build docker image with the operator.
	docker build -t ${IMG} -f Dockerfile $(PROJECT_ROOT)

.PHONY: docker-push
docker-push: ## Push docker image with the operator.
	docker push ${IMG}

##@ Deployment
IGNORE_NOT_FOUND = false
OPERATOR_NAME = dockerregistry-operator

.PHONY: build-image-operator
build-image-operator: ## Build dockerregistry operator from local sources on k3d
	docker build -t $(OPERATOR_NAME) -f Dockerfile $(PROJECT_ROOT)

.PHONY: install-operator-k3d
install-operator-k3d: build-image-operator ## Build and install dockerregistry operator from local sources on k3d
	$(eval HASH_TAG=$(shell docker images $(OPERATOR_NAME):latest --quiet))
	docker tag $(OPERATOR_NAME) $(OPERATOR_NAME):$(HASH_TAG)

	k3d image import $(OPERATOR_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment dockerregistry-operator -n kyma-system manager=$(OPERATOR_NAME):$(HASH_TAG)

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR)/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with IGNORE_NOT_FOUND=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR)/crd | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

.PHONY: deploy-main
deploy-main: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl create namespace kyma-system || true
	$(KUSTOMIZE) build $(CONFIG_OPERATOR) | kubectl apply -f -

.PHONY: generate-kustomization-dev
generate-kustomization-dev:
	cp $(CONFIG_OPERATOR_DEV)/kustomization.yaml.tpl $(CONFIG_OPERATOR_DEV)/kustomization.yaml

.PHONY: deploy-release
deploy-release: ## Deploy operator to the k8s cluster specified in ~/.kube/config with image from latest release.
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml

.PHONY: deploy
deploy: manifests kustomize generate-kustomization-dev ## Deploy controller to the K8s cluster specified in ~/.kube/config with image from IMG env.
	kubectl create namespace kyma-system || true
	cd $(CONFIG_OPERATOR_DEV) && $(KUSTOMIZE) edit set image europe-docker.pkg.dev/kyma-project/prod/dockerregistry-operator=${IMG}
	$(KUSTOMIZE) build $(CONFIG_OPERATOR_DEV) | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with IGNORE_NOT_FOUND=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR) | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

.PHONY: render-manifest
render-manifest: kustomize ## Render dockerregistry-operator.yaml manifest.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR) > $(PROJECT_ROOT)/dockerregistry-operator.yaml

.PHONY: apply-default-dockerregistry-cr
apply-default-dockerregistry-cr: ## Apply the k3d dockerregistry CR.
	kubectl apply \
		-f ${PROJECT_ROOT}/config/samples/default-dockerregistry-cr.yaml

.PHONY: remove-dockerregistry
remove-dockerregistry: ## Remove Dockerregistry CR
	kubectl delete dockerregistry -n kyma-system default --timeout 2m || (kubectl get dockerregistry -n kyma-system -oyaml && false)

.PHONY: check-dockerregistry-installation
check-dockerregistry-installation: ## Wait for Dockerregistry CR to be in Ready state.
	# wait some time to make sure operator starts the reconciliation first
	sleep 10

	./hack/verify_dockerregistry_status.sh || \
		(make print-dockerregistry-details && false)

	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry-operator --timeout=60s || \
		(make print-dockerregistry-details && false)

.PHONY: print-dockerregistry-details
print-dockerregistry-details: ## Print all pods, deploys and dockerregistry CRs in the kyma-system namespace.
	kubectl get dockerregistry -n kyma-system -oyaml
	kubectl get deploy -n kyma-system -oyaml
	kubectl get pods -n kyma-system -oyaml

##@ Module
.PHONY: module-image
module-image: docker-build docker-push ## Build the Module Image and push it to a registry defined in IMG.
	echo "built and pushed module image $(IMG)"

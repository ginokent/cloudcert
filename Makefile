SHELL := /bin/bash
GITROOT := $(shell git rev-parse --show-toplevel)
PRE_PUSH := ${GITROOT}/.git/hooks/pre-push
VERSION := $(shell git describe --tags --abbrev=0 --always)
REVISION := $(shell git log -1 --format='%H')
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
TIMESTAMP := $(shell git log -1 --format='%cI')

GOLANGCI_VERSION := 1.46.2

IMAGE_TAG := $(shell git describe --always --long --tags --dirty=-dirty$$(echo $$(git diff --cached; git diff) | sha1sum | head -c 7))
IMAGE_LOCAL := cloudacme:${IMAGE_TAG}
IMAGE_REMOTE := asia-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/newtstat/cloudacme:${IMAGE_TAG}

.DEFAULT_GOAL := help
.PHONY: help
help: githooks ## display this doc
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: githooks
githooks:  ## install githooks
	@test -f "${PRE_PUSH}" || cp -ai "${GITROOT}/.githooks/pre-push" "${PRE_PUSH}"

.PHONY: setup
setup: githooks ## setup
	GOBIN=${GITROOT}/.local/bin go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/envoyproxy/protoc-gen-validate

.PHONY: buf-mod-update
buf-mod-update: ## buf mod update
	cd ./proto && buf mod update

.PHONY: buf
buf: ## buf generate
	cd ./proto && buf generate

.PHONY: tidy
tidy:
	# tidy
	go mod tidy
	git diff --exit-code

.PHONY: lint
lint:  ## lint
	# lint
	# cf. https://github.com/golangci/golangci-lint/releases
	if [[ ! -x ./.local/bin/golangci-lint ]]; then GOBIN=${GITROOT}/.local/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v${GOLANGCI_VERSION}; fi
	# cf. https://golangci-lint.run/usage/linters/
	./.local/bin/golangci-lint run --fix --sort-results
	git diff --exit-code

.PHONY: test
test:  ## test
	# test
	go test -v -race -p=4 -parallel=8 -timeout=300s -cover -coverprofile=./coverage.txt ./...
	go tool cover -func=./coverage.txt

.PHONY: ci
ci: githooks tidy lint test ## ci

.PHONY: gobuild
gobuild: ## build
	go build -ldflags "-X github.com/newtstat/cloudacme/config.version=${VERSION} -X github.com/newtstat/cloudacme/config.revision=${REVISION} -X github.com/newtstat/cloudacme/config.branch=${BRANCH} -X github.com/newtstat/cloudacme/config.timestamp=${TIMESTAMP}" ./cmd/cloudacme/...

.PHONY: run
run: gobuild ## run
	ADDR=localhost GRPC_ENDPOINT=localhost:9090 ./cloudacme

.PHONY: air
air:  ## air
	if [[ ! -x ./.local/bin/air ]]; then curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b ${GITROOT}/.local/bin; fi
	ADDR=localhost GRPC_ENDPOINT=localhost:9090 air

.PHONY: GOOGLE_CLOUD_PROJECT
GOOGLE_CLOUD_PROJECT:
	@if [[ -z "${GOOGLE_CLOUD_PROJECT}" ]]; then printf "\033[1;31m%s\033[0m\n" "ERROR: GOOGLE_CLOUD_PROJECT is empty"; exit 1; fi

.PHONY: build
build:  ## docker build
	docker build -t ${IMAGE_LOCAL} --build-arg VERSION=${VERSION} --build-arg REVISION=${REVISION} --build-arg BRANCH=${BRANCH} --build-arg TIMESTAMP=${TIMESTAMP} .

.PHONY: push
push: GOOGLE_CLOUD_PROJECT ## docker push
	docker tag ${IMAGE_LOCAL} ${IMAGE_REMOTE}
	docker push ${IMAGE_REMOTE}

.PHONY: build-push
build-push: GOOGLE_CLOUD_PROJECT build push ## docker build push

.PHONY: gcloud-run-deploy
gcloud-run-deploy: GOOGLE_CLOUD_PROJECT ## gcloud run deploy
	gcloud --project="${GOOGLE_CLOUD_PROJECT}" run deploy cloudacme \
		--platform=managed \
		--max-instances=1 \
		--region=asia-northeast1 \
		--image=${IMAGE_REMOTE} \
		--set-env-vars "GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT}" \
		--set-env-vars "SPAN_EXPORTER=gcloud" \
		--set-env-vars "GRPC_ENDPOINT=localhost:9090"

.PHONY: build-push-gcloud-run-deploy
build-push-gcloud-run-deploy: GOOGLE_CLOUD_PROJECT build push gcloud-run-deploy ## docker build push gcloud run deploy

.PHONY: terraform-apply
terraform-apply: GOOGLE_CLOUD_PROJECT ## terraform apply
	GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT} TF_VAR_container_image=${IMAGE_REMOTE} terraform -chdir=terraform/gcloud apply

.PHONY: build-push-terraform-apply
build-push-terraform-apply: GOOGLE_CLOUD_PROJECT build push terraform-apply ## docker build push terraform apply

SHELL            := /bin/bash

.DEFAULT_GOAL := help
.PHONY: help
help: githooks ## display this help documents.
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: githooks
githooks:  ## githooks をインストールします。
	@test -f "${PRE_PUSH}" || cp -ai "${GITROOT}/.githooks/pre-push" "${PRE_PUSH}"

.PHONY: setup
setup: githooks ## protoc 周りのツール郡などをセットアップします。
	GOBIN=${GITROOT}/.local/bin go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/envoyproxy/protoc-gen-validate

.PHONY: buf-mod-update
buf-mod-update: ## buf mod update を実行します。
	buf --debug --verbose mod update

.PHONY: buf
buf: ## buf generate を実行します。
	buf --debug --verbose generate

.PHONY: lint
lint:  ## go mod tidy の後に golangci-lint を実行します。
	# tidy
	go mod tidy
	git diff --exit-code go.mod go.sum
	# lint
	# cf. https://github.com/golangci/golangci-lint/releases
	# if [[ ! -x ./.local/bin/golangci-lint ]]; then GOBIN=${GITROOT}/.local/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v${GOLANGCI_VERSION}; fi
	# cf. https://golangci-lint.run/usage/linters/
	./bin/golangci-lint run --fix --sort-results
	git diff --exit-code

.PHONY: test
test:  ## go test を実行し coverage を出力します。
	# test
	go test -v -race -p=4 -parallel=8 -timeout=300s -cover -coverprofile=./coverage.txt ./...
	go tool cover -func=./coverage.txt

.PHONY: ci
ci: lint test ## CI 上で実行する lint や test のコマンドセット

.PHONY: up
up:  ## docker compose up -d && docker compose logs -f
	docker compose up -d

.PHONY: down
down:  ## docker compose を終了します。 image や volume も削除します。
	docker compose down --rmi all --volumes --remove-orphans

.PHONY: restart
restart: down up ## docker compose を再起動します。

.PHONY: logs
logs:  ## docker compose のログを出力します。
	@printf '[\033[36mNOTICE\033[0m] %s\n' "プロンプトに戻るには Ctrl+C を押してください。"
	docker compose logs -f

.PHONY: gobuild
gobuild: ## go build を実行します。
	go build -ldflags "-X github.com/${OWNER_NAME}/${APP_NAME}/config.version=${VERSION} -X github.com/${OWNER_NAME}/${APP_NAME}/config.revision=${REVISION} -X github.com/${OWNER_NAME}/${APP_NAME}/config.branch=${BRANCH} -X github.com/${OWNER_NAME}/${APP_NAME}/config.timestamp=${TIMESTAMP}" ./cmd/${APP_NAME}/...

.PHONY: run
run: gobuild ## go build を実行し、コンパイル結果を起動します。
	ADDR=localhost GRPC_ENDPOINT=localhost:9090 ./${APP_NAME}

.PHONY: air
air:  ## air を起動します。
	if [[ ! -x ./.local/bin/air ]]; then curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b ${GITROOT}/.local/bin; fi
	ADDR=localhost GRPC_ENDPOINT=localhost:9090 air

.PHONY: build
build:  ## タグ ${IMAGE_LOCAL}:${IMAGE_TAG} に向けて docker build を実行します。
	docker build -t ${IMAGE_LOCAL}:${IMAGE_TAG} --build-arg VERSION=${VERSION} --build-arg REVISION=${REVISION} --build-arg BRANCH=${BRANCH} --build-arg TIMESTAMP=${TIMESTAMP} .

.PHONY: GOOGLE_CLOUD_PROJECT
GOOGLE_CLOUD_PROJECT:
	@if [[ -z "${GOOGLE_CLOUD_PROJECT}" ]]; then printf "\033[1;31m%s\033[0m\n" "ERROR: GOOGLE_CLOUD_PROJECT is empty"; exit 1; fi

.PHONY: push
push: GOOGLE_CLOUD_PROJECT ## タグ ${IMAGE_LOCAL}:${IMAGE_TAG} を ${IMAGE_REMOTE}:${IMAGE_TAG} として docker push します。
	docker tag ${IMAGE_LOCAL}:${IMAGE_TAG} ${IMAGE_REMOTE}:${IMAGE_TAG}
	docker push ${IMAGE_REMOTE}:${IMAGE_TAG}

.PHONY: build-push
build-push: GOOGLE_CLOUD_PROJECT build push ## make build の後に make push を実行します。

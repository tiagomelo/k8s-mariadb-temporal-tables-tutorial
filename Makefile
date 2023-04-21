SHELL = /bin/bash

MARIADB_BIN=kubectl exec -it deployment/mariadb -- mysql -uhr -phr123! -D hr

MARIADB_TEST_DOCKER_CONTAINER=temporal-tables-mariadb-test
MARIADB_TEST_BIN=docker exec -it $(MARIADB_TEST_DOCKER_CONTAINER) mysql -u$(MARIADB_USER) -p$(MARIADB_PASSWORD) -D$(MARIADB_DATABASE)

API_TAG=tiagomelo/hr-api:latest
MIGRATIONS_TAG=tiagomelo/migrations:latest

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: mariadb-console
## mariadb-console: launches mariadb local database console
mariadb-console:
	@ $(MARIADB_BIN)

.PHONY: mariadb-test-console
## mariadb-test-console: launches mariadb local test database console
mariadb-test-console: export MARIADB_DATABASE=hr
mariadb-test-console: export MARIADB_USER=hr
mariadb-test-console: export MARIADB_PASSWORD=hr123!
mariadb-test-console: export MARIADB_ROOT_PASSWORD=123456
mariadb-test-console: export MARIADB_HOST_NAME=localhost
mariadb-test-console: export MARIADB_PORT=3311
mariadb-test-console:
	@ $(MARIADB_TEST_BIN)

# ==============================================================================
# Docker images

.PHONY: build-api-img
## build-api-img: builds api image
build-api-img:
	@ eval $$(minikube -p minikube docker-env) ; \
	docker build --no-cache -t $(API_TAG) -f k8s/api/Dockerfile .

.PHONY: build-migrations-img
## build-migrations-img: builds migrations image
build-migrations-img:
	@ eval $$(minikube -p minikube docker-env) ; \
	docker build --no-cache -t $(MIGRATIONS_TAG) -f k8s/mariadb/Dockerfile .

# ==============================================================================
# Database migrations

.PHONY: migrate-setup
## migrate-setup: install golang-migrate
migrate-setup:
	@ if [ -z "$$(which migrate)" ]; then echo "Installing migrate command..."; go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest; fi

.PHONY: create-migrations
## create-migration: creates up and down migration files for a given name (make create-migrations NAME=<desired_name>)
create-migration: migrate-setup
	@ if [ -z "$(NAME)" ]; then echo >&2 please set the name of the migration via the variable NAME; exit 2; fi
	@ migrate create -ext sql -dir db/migrations -seq -digits 4 $(NAME)

.PHONY: migrate-db
## migrate-db: runs database migrations
migrate-db: build-migrations-img
	@ echo "Migrating MariaDB..."
	@ until $(MARIADB_BIN) -e 'SELECT 1' >/dev/null 2>&1 && exit 0; do \
	  >&2 echo "MariaDB not ready, sleeping for 5 secs..."; \
	  sleep 5 ; \
	done
	@ echo "... MariaDB is up and running!"
	@ echo "Applying database migrations..."
	@ kubectl apply -f k8s/mariadb/migrations.yaml
	@ echo "... done."

.PHONY: migrate-test-db
## migrate-test-db: runs database migrations into test db
migrate-test-db:
	@ echo "Setting up test MariaDB..."
	@ unset `env|grep DOCKER|cut -d\= -f1` ;\
	docker-compose up -d mariadb_test migrate_test
	@ until $(MARIADB_TEST_BIN) -e 'SELECT 1' >/dev/null 2>&1 && exit 0; do \
	  >&2 echo "MariaDB not ready, sleeping for 5 secs..."; \
	  sleep 5 ; \
	done
	@ echo "... MariaDB is up and running!"
	
# ==============================================================================
# Minikube

.PHONY: minikube-setup
## minikube-setup: starts minikube and build api and migration images
minikube-setup:
	@ minikube start ; \
	minikube addons enable kong ; \
	eval $$(minikube -p minikube docker-env) ; \
	docker build --no-cache -t $(MIGRATIONS_TAG) -f k8s/mariadb/Dockerfile . ; \
	docker build --no-cache -t $(API_TAG) -f k8s/api/Dockerfile .

# ==============================================================================
# Deployment

.PHONY: delete-api-deployment
## delete-api-deployment: deletes api deployment. Useful for redeploying the api
delete-api-deployment:
	@ kubectl delete deployment hr-api

.PHONY: deploy-api
## deploy-api: deploys the api
deploy-api:
	@ kubectl apply -f k8s/api/deployment.yaml && kubectl apply -f k8s/api/service.yaml

.PHONY: deploy-db
## deploy-db: deploys mariadb
deploy-db:
	@ kubectl apply -f k8s/mariadb/deployment.yaml && kubectl apply -f k8s/mariadb/service.yaml

.PHONY: deploy-test-db
## deploy-test-db: deploys mariadb test instance
deploy-test-db:
	@ kubectl apply -f k8s/mariadb/test_db_deployment.yaml && kubectl apply -f k8s/mariadb/service.yaml

.PHONY: redeploy-api
## redeploy-api: for redeploying the api
redeploy-api: delete-api-deployment build-api-img deploy-api

.PHONY: apply-ingress-rule
## apply-ingress-rule: applies the ingress rule
apply-ingress-rule:
	@ kubectl apply -f k8s/kong/ingress_rule.yaml

.PHONY: delete-ingress-rule
## delete-ingress-rule: deletes the ingress rule
delete-ingress-rule:
	@ kubectl delete -f k8s/kong/ingress_rule.yaml

# ==============================================================================
# kubectl

.PHONY: api-logs
## api-logs: display api's logs
api-logs:
	@ kubectl logs deployment/hr-api

.PHONY: pods
## pods: displays the running pods in Minikube at default namespace
pods:
	@ kubectl get pods

# ==============================================================================
# Kong

.PHONY: kong
## kong: on macbooks with m1 chip, we need to keep terminal open to run Kong
kong:
	@ minikube service -n kong kong-proxy --url | head -1

# ==============================================================================
# Cleanup
.PHONY: cleanup
## cleanup: cleans everything
cleanup:
	@ kubectl delete deployment mariadb
	@ kubectl delete service mariadb
	@ kubectl delete deployment mariadb-test
	@ kubectl delete service mariadb-test
	@ kubectl delete deployment hr-api
	@ kubectl delete service hr-api
	@ kubectl delete ingress hr-api-ingress

# ==============================================================================
# Tests

.PHONY: test
## test: runs unit tests
test:
	@ go test -cover -v ./... -count=1

.PHONY: coverage
## coverage: run unit tests and generate coverage report in html format
coverage:
	@ go test -coverprofile=coverage.out ./...  && go tool cover -html=coverage.out

.PHONY: int-test
## int-test: runs integration tests
int-test: export MARIADB_DATABASE=hr
int-test: export MARIADB_USER=hr
int-test: export MARIADB_PASSWORD=hr123!
int-test: export MARIADB_ROOT_PASSWORD=123456
int-test: export MARIADB_HOST_NAME=localhost
int-test: export MARIADB_PORT=3311
int-test: migrate-test-db
	@ go test -v ./test --tags=integration

# ==============================================================================
# Code quality

.PHONY: vet
## vet: runs go vet
vet:
	@ go vet ./...

.PHONY: lint
## lint: runs linter for all packages
lint:
	@ unset `env|grep DOCKER|cut -d\= -f1` ;\
	docker run --rm -v "`pwd`:/workspace:cached" -w "/workspace/." golangci/golangci-lint:latest golangci-lint run

.PHONY: vul-setup
## vul-setup: installs Golang's vulnerability check tool
vul-setup:
	@ if [ -z "$$(which govulncheck)" ]; then echo "Installing Golang's vulnerability detection tool..."; go install golang.org/x/vuln/cmd/govulncheck@latest; fi

.PHONY: vul-check
## vul-check: checks for any known vulnerabilities
vul-check: vul-setup
	@ govulncheck ./...

# ==============================================================================
# Swagger

.PHONY: swagger
## swagger: generates api's documentation
swagger: 
	@ unset `env|grep DOCKER|cut -d\= -f1` ;\
	docker run --rm -it -v $(HOME):$(HOME) -w $(PWD) quay.io/goswagger/swagger generate spec -o doc/swagger.json

.PHONY: swagger-ui
## swagger-ui: launches swagger ui
swagger-ui: swagger
	@ docker-compose up swagger-ui

# ==============================================================================
# App's execution

.PHONY: run
## run: does all needed setup and runs the api
run: minikube-setup deploy-api deploy-db migrate-db apply-ingress-rule
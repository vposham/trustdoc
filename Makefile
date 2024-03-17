# CONTAINER_ENGINE is set to either podman or docker. prefers podman over docker.
ifeq ($(shell command -v podman 2> /dev/null),)
	CONTAINER_ENGINE=docker
else
    CONTAINER_ENGINE=podman
endif

test:
	go clean -testcache && \
	go test -count=1 -race -shuffle=on `go list ./... | grep -v internal/db/sqlc/raw` -cover

runApp:
	export appEnv=local && go run main.go

killApp:
	lsof -i:8080 -Fp | head -n 1 | sed 's/^p//' | xargs kill

dbInstanceUp:
	${CONTAINER_ENGINE} run --name local-db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine && sleep 5

dbInstanceDown:
	${CONTAINER_ENGINE} stop local-db && ${CONTAINER_ENGINE} rm local-db

createDb:
	${CONTAINER_ENGINE} exec -it local-db createdb --username=root --owner=root doc_db

dropDb:
	${CONTAINER_ENGINE} exec -it local-db dropdb doc_db

migrateUp:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5432/doc_db?sslmode=disable" -verbose up

migrateDown:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5432/doc_db?sslmode=disable" -verbose down

resetDbData:
	${CONTAINER_ENGINE} exec -it local-db psql -U root -d postgres -c "DROP DATABASE IF EXISTS doc_db" && \
	make create-db && \
	make migrate-up && \
	make load-test-data-for-local

##########################################  SQLC GENERATE  ####################################################

sqlc:
	sqlc generate

##########################################  BLOB STORE - MINIO  ####################################################
minioUp:
	mkdir -p ./minioData && \
	${CONTAINER_ENGINE} run -p 9000:9000 -p 9001:9001 --name minio -v ./minioData:/data -e "MINIO_ROOT_USER=minio" -e "MINIO_ROOT_PASSWORD=CHANGEME123" quay.io/minio/minio server /data --console-address ":9001"

minioDown:
	${CONTAINER_ENGINE} stop minio && ${CONTAINER_ENGINE} rm minio

minioPurge:
	rm -rf ./minioData

##########################################  BUILD AND RUN ####################################################
start-all: dbInstanceUp \
		   createDb  \
		   migrateUp \
		   minioUp \
		   runApp

stop-all: killApp \
		  dbInstanceDown

##########################################  LINTING  ####################################################

# Make sure to get the owl-quality project into your local and have it in the same parent folder of this project
# install revive
runRevive:
	revive -config config.toml  -formatter friendly ./... ./... > lint-issues.txt


 # Install goimports-reviser for formatting imports/code.
 # Install from here - https://github.com/incu6us/goimports-reviser#install
 # imports are sorted in following order :
 # (1) go in-built packages (2) 3rd part packages (3) organization (od) packages and (4) local app packages
 # Code is formatted in same way as gofmt
importFormat:
	gofmt -w -s .
	goimports-reviser -rm-unused -set-alias -format -company-prefixes github.com/vposham -imports-order std,general,company,project -recursive ./...

vet:
	go vet ./...

lint: runRevive importFormat vet

##########################################  INTEGRATION/PERF TESTS  ####################################################

# Make sure app is running before running this
# Update basic auth credentials in integration/environment.json to support local integration testing
integrationTest:
	newman run --env-var host=http://localhost:8080 ./integration/integration.json -e ./integration/environment.json

# Make sure app is running before running this, increase duration from 10s to something like 3m during actual test
perf:
	k6 run -e TEST_IN_LOCAL=true -d 10s --vus 5 --insecure-skip-tls-verify ./performance/basicTest.js


##########################################  SECURITY FIXES  ####################################################

# ups updates all minor and security fixes - Running this resolves Security issues
ups: goups

goups:
	go get -t -u ./... && go mod tidy

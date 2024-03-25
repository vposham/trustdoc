# CRT (container runtime) is set to either podman or docker. prefers podman over docker.
ifeq ($(shell command -v podman 2> /dev/null),)
	CRT=docker
else
    CRT=podman
endif

test:
	go clean -testcache && \
	go test -count=1 -race -shuffle=on `go list ./... | grep -v internal/db/sqlc/raw | grep -v internal/bc/contracts` -cover

runApp:
	export appEnv=local && go run main.go

killApp:
	lsof -i:8080 -Fp | head -n 1 | sed 's/^p//' | xargs kill

dbInstanceUp:
	${CRT} run --name local-db -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine && sleep 5

dbInstanceDown:
	${CRT} rm -f local-db

createDb:
	${CRT} exec -it local-db createdb --username=root --owner=root doc_db

dropDb:
	${CRT} exec -it local-db dropdb doc_db

migrateUp:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5432/doc_db?sslmode=disable" -verbose up

migrateDown:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5432/doc_db?sslmode=disable" -verbose down

resetDbData:
	${CRT} exec -it local-db psql -U root -d postgres -c "DROP DATABASE IF EXISTS doc_db" && \
	make create-db && \
	make migrate-up && \
	make load-test-data-for-local

##########################################  SQLC GENERATE  ####################################################

sqlc:
	sqlc generate

##########################################  BLOB STORE - MINIO  ####################################################

minioUp:
	mkdir -p ./minioData && \
	${CRT} run -p 9000:9000 -p 9001:9001 --name minio -v ./minioData:/data -e "MINIO_ROOT_USER=minio" -e "MINIO_ROOT_PASSWORD=CHANGEME123" -d minio/minio server /data --console-address ":9001" && sleep 5

minioDown:
	${CRT} rm -f minio

minioBkt:
	${CRT} exec -it minio mc alias set myminio http://localhost:9000 minio CHANGEME123 && \
	${CRT} exec -it minio mc mb myminio/docs-store && \
	${CRT} exec -it minio mc anonymous set public myminio/docs-store

minioPurge:
	rm -rf ./minioData

##########################################  BUILD AND RUN ####################################################

startAll: dbInstanceUp \
		   createDb  \
		   migrateUp \
		   minioUp \
		   minioBkt \
		   runApp

stopAll: killApp \
		 minioDown \
		 dbInstanceDown \
		 minioPurge

infra: dbInstanceUp \
		   createDb  \
		   migrateUp \
		   minioUp \
		   minioBkt

##########################################  LINTING  ####################################################

# Make sure to get the owl-quality project into your local and have it in the same parent folder of this project
# install revive
runRevive:
	revive -config revive.toml  -formatter friendly ./... ./... > lint-issues.txt


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
integrationTest:
	newman run --env-var host=http://localhost:8080 ./docs/trustdoc.postman_collection.json

# Make sure app is running before running this, increase duration from 10s to something like 3m during actual test
perf:
	k6 run -e -d 10s --vus 5 --insecure-skip-tls-verify ./docs/k6_test.js

##########################################  SECURITY FIXES  ####################################################

# ups updates all minor and security fixes - Running this resolves Security issues
ups: goUps npmUps

goUps:
	go get -t -u ./... && go mod tidy

npmUps:
	npm i -g npm-check-updates && npm i

# npmUpsMjr updates all deps including major changes which needs code changes.
npmUpsMjr:
	npm i -g npm-check-updates && ncu -u && npm i

##########################################  COMPILE AND GO GEN CONTRACTS  ####################################################

sol2go: solc abigen cln

solc:
	solc --base-path ./ --include-path node_modules --abi internal/bc/contracts/DocumentToken.sol -o internal/bc/contracts --overwrite

abigen:
	abigen --abi internal/bc/contracts/DocumentToken.abi --pkg bc --type DocumentToken --out internal/bc/DocumentToken.go

cln:
	rm -r internal/bc/contracts/*.abi

solccompile:
	solc --base-path ./ --combined-json bin,bin-runtime,srcmap,srcmap-runtime,abi,userdoc,devdoc,metadata  --optimize --evm-version berlin --allow-paths . --include-path ./node_modules ./internal/bc/contracts/DocumentToken.sol

##########################################  CONTAINERIZE  ####################################################

build:
	${CRT} build -t trustdoc .

dup:
	docker-compose up

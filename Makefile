GITHASH=`git log -1 --pretty=format:"%h" || echo "???"`
CURDATE=`date -u +%Y%m%d%H%M%S`
VERSION=v0.1.1

APPVERSION=${VERSION}_${GITHASH}_${CURDATE}

init:
	go install -v ./vendor/github.com/gogo/protobuf/protoc-gen-gogofast
	# go install -v ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install -v ./vendor/gopkg.in/reform.v1/reform
	go install -v ./vendor/github.com/mwitkow/go-proto-validators/protoc-gen-govalidators

install:
	go install -v ./...
	go test -i ./...

test-short: install
	go test -v -short ./...

test: install
	go test -v -count 1 -race -short ./...
	go test -v -count 1 -race -timeout 30m ./tests \
		--consul-key eventools-settings \
		--pg-files-path "${PWD}/scripts/postgres_schema/*.sql" \
		--run=

setup:
	go test -v -count 1 -race -timeout 30m ./tests \
		--only-setup \
		--consul-key eventools-settings \
		--pg-files-path "${PWD}/scripts/postgres_schema/*.sql"

gen:
	# protobuf / gRPC
	find ./api -name '*.pb.go' -delete
	protoc --proto_path=. --proto_path=./vendor --govalidators_out=. --gogofast_out=plugins=grpc:. ./api/yourapp_api/*.proto

	# reform
	find ./services -name '*_reform.go' -delete
	go generate ./services/...

up:
	docker-compose up --detach --force-recreate --renew-anon-volumes --remove-orphans

down:
	docker-compose down --volumes --remove-orphans

build-race:
	go build -v -race -o ./bin/yourapp-debug ./cmd/yourapp/main.go

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -ldflags "-X main.VERSION=${APPVERSION}" -o ./bin/yourapp ./cmd/yourapp/main.go

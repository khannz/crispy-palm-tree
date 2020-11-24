export PATH := ./bin:$(PATH)
export GO111MODULE := on

dummy-dl-mods:
	cd t1-dummy && go mod download

dummy-grpc:
	mkdir -p ./t1-dummy/grpc-dummy
	mkdir -p ./t1-dummy/grpc-orch
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-dummy/ --go-grpc_out=./t1-dummy/grpc-dummy/ ./proto/dummy.proto
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-orch/ --go-grpc_out=./t1-dummy/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
dummy-rpm-snapshot:
	cd t1-dummy && goreleaser --snapshot --skip-publish --rm-dist

dummy-bin:
	cd t1-dummy && CGO_ENABLED=0 go build -o lbost1ad

dummy-clean:
	rm -rf ./t1-dummy/dist
	rm -f ./t1-dummy/lbost1ad

healthcheck-dl-mods:
	cd t1-healthcheck && go mod download

healthcheck-grpc:
	mkdir -p ./t1-healthcheck/grpc-healthcheck
	protoc -I ./proto/ --go_out=./t1-healthcheck/grpc-healthcheck/ --go-grpc_out=./t1-healthcheck/grpc-healthcheck/ ./proto/healthcheck.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
healthcheck-rpm-snapshot:
	cd t1-healthcheck && goreleaser --snapshot --skip-publish --rm-dist

healthcheck-bin:
	cd t1-healthcheck && CGO_ENABLED=0 go build -o lbost1ah

healthcheck-clean:
	rm -rf ./t1-healthcheck/dist
	rm -f ./t1-healthcheck/lbost1ad
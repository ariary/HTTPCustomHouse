before.build:
	go mod download

build.httpcustomhouse:
	@echo "build in ${PWD}";go build -o httpcustomhouse cmd/httpcustomhouse/main.go

build.httpoverride:
	@echo "build in ${PWD}";go build -o httpoverride cmd/httpoverride/main.go

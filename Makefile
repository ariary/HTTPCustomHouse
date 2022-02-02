before.build:
	go mod download

build.httpcustomhouse:
	@echo "build in ${PWD}";go build -o httpcustomhouse cmd/httpcustomhouse/main.go


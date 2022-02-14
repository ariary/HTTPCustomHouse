before.build:
	go mod tidy && go mod download

build.httpcustomhouse:
	@echo "build in ${PWD}";go build -o httpcustomhouse cmd/httpcustomhouse/main.go

build.httpoverride:
	@echo "build in ${PWD}";go build -o httpoverride cmd/httpoverride/main.go

build.httpclient:
	@echo "build in ${PWD}";go build -o httpclient cmd/httpclient/main.go
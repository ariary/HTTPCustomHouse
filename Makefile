before.build:
	go mod download && go mod vendor

build.customOfficer-CL:
	@echo "build in ${PWD}";go build -o customOfficer-CL cmd/customOfficer-CL/main.go

build.customOfficer-TE:
	@echo "build in ${PWD}";go build -o customOfficer-TE cmd/customOfficer-TE/main.go

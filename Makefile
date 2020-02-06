now := $(shell date -R)
sha := $(shell git rev-parse HEAD)

.PHONY: all serve test coverage

all:
	go build -ldflags "-X main.gitVer=$(sha) -X \"main.buildTime=$(now)\""

serve:
	go build -ldflags "-X main.gitVer=$(sha) -X \"main.buildTime=$(now)\""
	./mibig-api serve --debug

test:
	go test ./...

coverage:
	go test -coverprofile=cover.prof ./...
	go tool cover -html=cover.prof

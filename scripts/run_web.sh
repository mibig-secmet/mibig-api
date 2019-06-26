#!/bin/bash

now=$(date -R)
sha=$(git rev-parse HEAD)

go run -ldflags "-X main.gitVer=${sha} -X \"main.buildTime=${now}\"" ./cmd/web $@

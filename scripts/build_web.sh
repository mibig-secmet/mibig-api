#!/bin/bash

now=$(date -R)
sha=$(git rev-parse HEAD)

go build -ldflags "-X main.gitVer=${sha} -X \"main.buildTime=${now}\"" ./cmd/web

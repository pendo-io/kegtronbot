#!/bin/bash

go mod vendor
go build -mod=vendor ./...
gcloud --project pendo-pankbot app deploy internal/app/backend.yaml --quiet --promote --no-stop-previous-version

#!/bin/bash

go mod vendor
go build -mod=vendor ./...
gcloud --project pendo-kegtronbot-test app deploy internal/app/app.yaml --quiet --promote --no-stop-previous-version

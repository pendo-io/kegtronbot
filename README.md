# Pankbot

### Purpose

* Creating and viewing panks

### Uses

* `/pank` interactively create a pank
* `/pank [:pank-emoji:] [@user] [reason] [<private(ly)>]` create a pank
* `/pank me` what panks you have sent and received (your eyes only)
* `/pank report` display a report of all *public* panks (your eyes only)
* `/pank help`  

#### First use

Clone this repo (If you are reading this, you probably already know where to go, but just in case you hit your head, head over to https://github.com/pendo-io/pankbot)

Update the submodules

```shell
go mod vendor
```

Build
```shell
go build -mod=vendor ./...
```

#### Run
Running this will be done primarily by deploying the project to GAE and using it through the pendo-test slack.

#### Testing
You can run the unit tests by running `go test -tags appenginevm -v  ./...`

Use a look-aside version in GAE and test on the pendo-test slack
```shell
sh ./deploy-test.sh
```

#### Deploy

```shell
sh ./deploy-prod.sh
```

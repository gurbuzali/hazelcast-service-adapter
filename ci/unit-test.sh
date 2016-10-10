#!/bin/bash

set -ex

export GOPATH=$PWD

go version
go test ./...
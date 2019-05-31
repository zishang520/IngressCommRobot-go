#!/usr/bin/env bash
export GOPATH=`pwd`
go get github.com/PuerkitoBio/goquery
go get github.com/mattn/go-sqlite3
go get github.com/zishang520/persistent-cookiejar
go get golang.org/x/net/proxy
GOOS=darwin GOARCH=386 go build -ldflags "-s -w" -o bin/darwin_386/Ingress Ingress.go
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/darwin_amd64/Ingress Ingress.go
GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o bin/linux_386/Ingress Ingress.go
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/linux_amd64/Ingress Ingress.go
GOOS=linux GOARCH=arm go build -ldflags "-s -w" -o bin/linux_arm/Ingress Ingress.go
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/linux_arm64/Ingress Ingress.go
GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o bin/windows_386/Ingress.exe Ingress.go
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/windows_amd64/Ingress.exe Ingress.go
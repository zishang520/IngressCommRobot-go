@echo OFF
set GOPATH="%~dp0"

echo ========================
echo build darwin_386/Ingress
set GOOS=darwin
set GOARCH=386
go build -o bin/darwin_386/Ingress Ingress.go

echo ========================
echo build darwin_amd64/Ingress
set GOOS=darwin
set GOARCH=amd64
go build -o bin/darwin_amd64/Ingress Ingress.go

echo ========================
echo build linux_386/Ingress
set GOOS=linux
set GOARCH=386
go build -o bin/linux_386/Ingress Ingress.go

echo ========================
echo build linux_amd64/Ingress
set GOOS=linux
set GOARCH=amd64
go build -o bin/linux_amd64/Ingress Ingress.go

echo ========================
echo build linux_arm/Ingress
set GOOS=linux
set GOARCH=arm
go build -o bin/linux_arm/Ingress Ingress.go

echo ========================
echo build linux_arm64/Ingress
set GOOS=linux
set GOARCH=arm64
go build -o bin/linux_arm64/Ingress Ingress.go

echo ========================
echo build windows_386/Ingress.exe
set GOOS=windows
set GOARCH=amd64
go build -o bin/windows_386/Ingress.exe Ingress.go

echo ========================
echo build windows_amd64/Ingress.exe
set GOOS=windows
set GOARCH=amd64
go build -o bin/windows_amd64/Ingress.exe Ingress.go

echo ========================
echo Successful
echo ========================
pause
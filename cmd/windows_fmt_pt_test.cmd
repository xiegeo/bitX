@echo off & setlocal enableextensions enabledelayedexpansion

pushd "%~dp0"

cd "..\src\network"
echo network
go fmt
protoc --go_out=. *.proto
call :go_test

cd "..\hashtree"
echo hashtree
go fmt
call :go_test

cd "..\server"
echo server
go fmt
call :go_test

ENDLOCAL & goto:EOF

:go_test
go test 1> nul || go test
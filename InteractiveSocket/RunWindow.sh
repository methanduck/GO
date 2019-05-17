#!/bin/sh
cd InteractiveSocket/Starter
go get
go build main.go
./main -addr "192.168.0.50" -port "6974"
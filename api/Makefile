#！/usr/bin/env bash
pwd = $(shell pwd)
parent = $(abspath $(pwd)/../)
gitProject = snakeAndLadder
#if [ -n "$1"]; then target=$1; fi
all:
	cd ../../ && for file in `find ./$(gitProject) -name '*.proto' -print`; \
	do echo $$file; protoc --proto_path=. --go_out=plugins=grpc:. $$file; done\

clean:
	find -name "*.pb.go" -delete

gateway:
	cd ../../ && for file in `find ./$(gitProject) -name '*.proto' -print`; \
    	do echo $$file; protoc --proto_path=. --go_out=plugins=grpc:. $$file; done\

swagger:

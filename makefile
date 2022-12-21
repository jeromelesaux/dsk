CC=go
RM=rm
MV=mv


SOURCEDIR=./cli
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
BINARIES=binaries

VERSION:=$(shell grep -m1 "appVersion" cli/main.go | sed 's/[", ]//g' | cut -d= -f2)
suffix=$(shell grep -m1 "appVersion" cli/main.go | sed 's/[", ]//g' | cut -d= -f2 | sed 's/[0-9.]//g')
snapshot=$(shell date +%FT%T)

ifeq ($(suffix),rc)
	appversion=$(VERSION)$(snapshot)
else 
	appversion=$(VERSION)
endif 

.DEFAULT_GOAL:=build


build: 
	(make clean)
	(make init)
	@echo "Update packages"
	go get -d ./...
	@echo "Compilation for linux amd64"
	(make compile ARCH=amd64 OS=linux)
	@echo "Compilation for windows amd64"
	(make compile ARCH=amd64 OS=windows EXT=.exe)
	@echo "Compilation for macos"
	(make compile ARCH=amd64 OS=darwin)
	@echo "Compilation for raspberry pi Raspbian 64bits"
	(make compile ARCH=arm64 OS=linux)
	@echo "Compilation for raspberry pi Raspbian 32bits"
	(make compile ARCH=arm OS=linux GOARM=5)
	@echo "Compilation for older windows"
	(make compile ARCH=386 OS=windows EXT=.exe)

init: 
	mkdir ${BINARIES}

clean:
	@echo "Cleaning project"
	rm -fr ${BINARIES}/

compile:
	GOOS=${OS} GOARCH=${ARCH} go build ${LDFLAGS} -o ${BINARIES}/dsk-${OS}-${ARCH}${EXT} $(SOURCEDIR)/main.go 
	zip ${BINARIES}/dsk-$(appversion)-${OS}-${ARCH}.zip ${BINARIES}/dsk-${OS}-${ARCH}${EXT}

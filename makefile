CC=go
RM=rm
MV=mv


SOURCEDIR=cli
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

VERSION:=$(shell grep -m1 "version" cli/main.go | sed 's/[", ]//g' | cut -d= -f2)
suffix=$(shell grep -m1 "version" cli/main.go | sed 's/[", ]//g' | cut -d= -f2 | sed 's/[0-9.]//g')
snapshot=$(shell date +%FT%T)

ifeq ($(suffix),rc)
	appversion=$(VERSION)$(snapshot)
else 
	appversion=$(VERSION)
endif 

.DEFAULT_GOAL:=build


build: 
	@echo "Compilation for linux"
	GOOS=linux go build ${LDFLAGS} -o dsk $(SOURCEDIR)/main.go
	zip dsk-$(appversion)-linux.zip dsk 
	@echo "Compilation for windows"
	GOOS=windows go build ${LDFLAGS} -o dsk.exe $(SOURCEDIR)/main.go
	zip dsk-$(appversion)-windows.zip dsk.exe 
	@echo "Compilation for macos"
	GOOS=darwin go build ${LDFLAGS} -o dsk $(SOURCEDIR)/main.go
	zip dsk-$(appversion)-macos.zip dsk  
	@echo "Compilation for raspberry pi Raspbian"
	GOOS=linux GOARCH=arm GOARM=5 go build ${LDFLAGS} -o dsk $(SOURCEDIR)/main.go 
	zip dsk-$(appversion)-arm.zip dsk  
	@echo "Compilation for older macos"
	GOOS=darwin GOARCH=386 go build ${LDFLAGS} -o dsk $(SOURCEDIR)/main.go 
	zip dsk-$(appversion)-macos-older.zip dsk  
	@echo "Compilation for older windows"
	GOOS=windows GOARCH=386 go build ${LDFLAGS} -o dsk.exe $(SOURCEDIR)/main.go 
	zip dsk-$(appversion)-windows-older.zip dsk  

clean:
	@echo "Cleaning project"
	rm -f dsk.exe
	rm -f dsk
	rm dsk-*.zip

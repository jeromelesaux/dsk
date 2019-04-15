CC=go
RM=rm
MV=mv


SOURCEDIR=./cli/
SOURCES := $(shell find . -name '*.go')
GOOS=linux
GOARCH=amd64

#VERSION:=$(shell grep -m1 "dskVersion string" *.go | sed 's/[", ]//g' | cut -d= -f2)
APP=dsk

BUILD_TIME=`date +%FT%T%z`
PACKAGES := 
DSKAPP=$(shell git rev-parse --short HEAD)

LIBS=

LDFLAGS=-ldflags  #"-X main.dskGitHash=$(DSKAPP)"

.DEFAULT_GOAL:=test

help:
		@echo ""
		@echo "***********************************************************"
		@echo "******** makefile's help, possible actions: ***************"
		@echo "*** test : execute test on the project"
		@echo "*** package : package the application"
		@echo "*** coverage-display : execute tests and display code coverage in your navigator"
		@echo "*** coverage : execute test and gets the code coverage on the project"
		@echo "*** fmt : execute go fmt on the project"
		@echo "*** audit : execute static audit on source code."
		@echo "*** deps : get the dependencies of the project"
		@echo "*** init : initialise the project"
		@echo "*** clean : clean binaries and project structure"
		@echo "*** package-zip : create the zip archive to delivery"
		@echo "***********************************************************"
		@echo ""


package: ${APP}
		@tar -cvzf ${APP}-${GOOS}-${GOARCH}.tar.gz ${APP}
		@echo "    Archive ${APP}-${GOOS}-${GOARCH}.tar.gz created"

test: $(APP) 
		@GOOS=${GOOS} GOARCH=${GOARCH} go test -cover ./...
		@echo " Tests OK."

coverage-display:
		$(shell ls -d */ | while read d; do d_=`echo $$d | tr -d '/'`;go test -coverprofile=$$d_.out github.com/jeromelesaux/dsk/$$d_; go tool cover -html=$$d_.out; done 1>/dev/null)
		@echo " Tests executed and results will be displayed in your navigator."

coverage: 
		@go get github.com/axw/gocov/gocov
		@gocov test ./... | gocov report

$(APP): fmt $(SOURCES)
		@echo "    Compilation des sources ${BUILD_TIME}"
		@echo ""
		@GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${APP} $(SOURCEDIR)/main.go
		@echo "    ${APP} generated."

build:
		@GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${APP} $(SOURCEDIR)/main.go
		@echo "    ${APP} generated."

fmt: audit
		@echo "    Go FMT"
		@$(foreach element,$(SOURCES),go fmt $(element);)

audit: deps
		@go tool vet -all . 2> audit.log &
		@echo "    Audit effectue"

deps: init
		@echo "    Download packages"		
		dep ensure -update -v

init: clean
		@echo "    Init of the project"
		@echo "    Version :: ${VERSION}"

clean:
		@if [ -f "${APP}" ] ; then rm ${APP} ; fi
		@if [ -f "${APP}-linux-amd64.tar.gz" ] ; then rm ${APP}-linux-amd64.tar.gz ; fi
		@rm -f *.out
		@echo "    Nettoyage effectuee"


execute: 
		@./${APP} -port 8080 -dbconfig dbconfig.json -appconfig appconfig.json -logerrorfile test.log -loglevel DEBUG

package-zip:  ${APP}
		@zip -r ${APP}-${GOOS}-${GOARCH}.zip ./${APP}
		@echo "    Archive ${APP}-${GOOS}-${GOARCH}.zip created"


#----------------------------------------------------------------------#
#----------------------------- docker actions -------------------------#
#----------------------------------------------------------------------#

DOCKER_IP=$(shell if [ -z "$(DOCKER_MACHINE_NAME)" ]; then echo 'localhost'; else docker-machine ip $(DOCKER_MACHINE_NAME); fi)

dockerBuild:
	docker build -t ${DSKAPP} .

dockerClean:
	docker rmi -force ${DSKAPP} .

dockerUp:
	docker-compose up -d

dockerStop:
	docker-compose stop
	docker-compose kill
	docker-compose rm -f

dockerBuildUp: dockerStop dockerBuild 

dockerWatch:
	@watch -n1 'docker ps | grep ${DSKAPP}'

dockerLogs:
	docker-compose logs -f

dockerBash:
	docker exec -it $(shell docker ps -aqf "name=${DSKAPP}" 2> /dev/null) bash

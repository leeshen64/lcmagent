UNAME := $(shell uname)
BINDIR := bin
EXE := cpe_agent

ifeq (${UNAME}, Darwin)
	BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S)
else
	BUILD_TIME=$(shell date --rfc-3339=seconds)
endif

all:
	make clean
	mkdir -p ${BINDIR}
	go build -a -ldflags "-s -w -X \"main.BUILD_TIME=${BUILD_TIME}\"" -o ${BINDIR}/${UNAME}/${EXE}

clean:
	rm -Rf ${BINDIR}

darwin:
	make clean
	mkdir -p ${BINDIR}/darwin
	GOOS=darwin GOARCH=amd64 go build -a -ldflags "-s -w -X \"main.BUILD_TIME=${BUILD_TIME}\"" -o ${BINDIR}/darwin/${EXE}

ubuntu:
	make clean
	mkdir -p ${BINDIR}/ubuntu
	GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w -X \"main.BUILD_TIME=${BUILD_TIME}\"" -o ${BINDIR}/ubuntu/${EXE}

arm.v5:
	mkdir -p ${BINDIR}/arm.v5
	GOOS=linux GOARCH=arm GOARM=5 go build -a -ldflags "-s -w -X \"main.BUILD_TIME=${BUILD_TIME}\"" -o ${BINDIR}/arm_v5/${EXE}

arm.v7:
	mkdir -p ${BINDIR}/arm.v7
	GOOS=linux GOARCH=arm GOARM=7 go build -a -ldflags "-s -w -X \"main.BUILD_TIME=${BUILD_TIME}\"" -o ${BINDIR}/arm_v7/${EXE}

mipsle:
	mkdir -p ${BINDIR}/mipsle
	GOOS=linux GOARCH=mipsle go build -a -ldflags "-s -w -X \"main.BUILD_TIME=${BUILD_TIME}\"" -o ${BINDIR}/mipsle/${EXE}

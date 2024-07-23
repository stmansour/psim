TOP=../..
BINDIR=${TOP}/dist/plato
THISDIR="apps/simtalk"
TEST_FAILURE_FILE=fail
THISDIR := $(notdir $(PWD))
BUILD_TIME := $(shell date "+%Y%m%dT%H%M%S")

sim: *.go
	go vet
	golint
	staticcheck
	go build -ldflags "-X 'github.com/stmansour/psim/util.buildID=$(BUILD_TIME)'" -o simtalk
	@echo "*** $(THISDIR): completed $(THISDIR) ***"

clean:
	go clean
	rm -rf fail simtalk
	@echo "*** ${THISDIR}: completed clean ***"

coverage:
	go tool cover -html=coverage.out 

test:
	@echo "*** ${THISDIR}: completed package ***"

package:
	mkdir -p ${BINDIR}/
	cp simtalk ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

release:
	cp simtalk /usr/local/plato/bin/

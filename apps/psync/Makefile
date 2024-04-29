TOP=../..
BINDIR=${TOP}/dist/platosim
THISDIR="apps/psync"
TEST_FAILURE_FILE=fail
THISDIR := $(notdir $(PWD))
BUILD_TIME := $(shell date "+%Y-%m-%d_%H:%M:%S")

psync: *.go
	go vet
	golint
	staticcheck
	go build -ldflags "-X 'github.com/stmansour/psim/util.buildID=$(BUILD_TIME)'" -o psync
	@echo "*** $(THISDIR): completed $(THISDIR) ***"

clean:
	go clean
	rm -rf fail psync config.json5 extres.json5 coverage.out
	@echo "*** ${THISDIR}: completed clean ***"

coverage:
	go tool cover -html=coverage.out 

test: config
	@touch $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out -coverpkg=../../... && rm -f ${TEST_FAILURE_FILE}
	@echo "*** ${THISDIR}: completed package ***"

config:
	cp ${TOP}/util/config.json5 .
	cp ~/.vault/extres.json5 .

package:
	mkdir -p ${BINDIR}/
	cp psync ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

TOP=..
THISDIR="util"

util: *.go
	@touch fail
	go vet
	if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	go install
	@rm -f fail
	@echo "*** ${THISDIR}: completed compile ***"

clean:
	@echo "*** ${THISDIR}: completed clean ***"

install:
	go install

test:
	@touch fail
	@go test github.com/stmansour/psim/util && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"
	@rm -f fail

package:
	mkdir -p ${TOP}/dist
	cp config.json5 ${TOP}/dist/
	@echo "*** ${THISDIR}: completed package ***"

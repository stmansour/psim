TOP=..
BINDIR=${TOP}/dist
THISDIR="core"
TEST_FAILURE_FILE = .tests_failed

core: *.go
	go vet
	@if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	go install
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -rf *.out
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/core && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

coverage:
	go tool cover -html=coverage.out

package:
	@echo "*** ${THISDIR}: completed package ***"

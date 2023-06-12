TOP=..
BINDIR=${TOP}/dist
THISDIR="tools"
TEST_FAILURE_FILE = .tests_failed

tools: *.go
	go vet
	@if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	mv tools csvfixer
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -rf csvfixer *.out
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@echo "*** ${THISDIR}: completed test ***"

coverage:
	go tool cover -html=coverage.out

package:
	mkdir -p ${BINDIR}
	cp csvfixer ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

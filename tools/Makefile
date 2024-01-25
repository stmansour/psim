TOP=..
BINDIR=${TOP}/dist/platosim
THISDIR="tools"
TEST_FAILURE_FILE=fail

tools: *.go
	go vet
	golint
	staticcheck
	go build
	mv tools csvfixer
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -rf csvfixer *.out fail
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/tools && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

coverage:
	go tool cover -html=coverage.out

package:
	mkdir -p ${BINDIR}
	cp csvfixer ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

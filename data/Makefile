TOP=..
BINDIR=${TOP}/dist/platosim
THISDIR=data
TEST_FAILURE_FILE=fail

data: *.go
	go vet
	@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	staticcheck
	go build
	go install
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	@rm -rf *.out fail
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/data && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

coverage:
	go tool cover -html=coverage.out

package:
	mkdir -p ${BINDIR}/data
	cd data;cp platodb.csv ../${BINDIR}/data;cd ..
	@echo "*** ${THISDIR}: completed package ***"

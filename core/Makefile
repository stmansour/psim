TOP=..
BINDIR=${TOP}/dist/platosim
THISDIR="core"
TEST_FAILURE_FILE=fail

core: *.go
	go vet
	@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	staticcheck
	go build
	go install
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -rf *.out fail data config.json5
	@echo "*** ${THISDIR}: completed clean ***"

test: config.json5 data/platodb.csv
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/core && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

config.json5:
	cp ${TOP}/util/config.json5 .

data/platodb.csv:
	mkdir -p data;cp ${TOP}/data/data/platodb.csv data/

coverage:
	go tool cover -html=coverage.out

package:
	@echo "*** ${THISDIR}: completed package ***"

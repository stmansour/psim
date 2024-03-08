TOP=..
BINDIR=${TOP}/dist/platosim
THISDIR="newcore"
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
	rm -rf data *.out fail config.json5 extres.json5
	@echo "*** ${THISDIR}: completed clean ***"

test: config db
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/${THISDIR} && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

config:
	cp ${TOP}/util/config.json5 .
	cp ${TOP}/util/extres.json5 .

db:
	mkdir -p data;cp ${TOP}/newdata/data/platodb.csv data/
	cp ${TOP}/newdata/data/misubclasses.csv data/

coverage:
	go tool cover -html=coverage.out

package:
	mkdir -p ${BINDIR}/data
	cp data/misubclasses.csv ${BINDIR}/data/
	@echo "*** ${THISDIR}: completed package ***"

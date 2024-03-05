TOP=..
BINDIR=${TOP}/dist/platosim
THISDIR=newdata
TEST_FAILURE_FILE=fail
DIRS=csvtosql

data: *.go
	go vet
	@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	staticcheck
	go build
	go install
	for dir in $(DIRS); do make -C $$dir;done
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	@rm -rf *.out fail
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/${THISDIR} && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	for dir in $(DIRS); do make -C $$dir test;done
	@echo "*** ${THISDIR}: completed test ***"

coverage:
	go tool cover -html=coverage.out
	for dir in $(DIRS); do make -C $$dir coverage;done

package:
	mkdir -p ${BINDIR}/data
	cd data;cp platodb.csv ../${BINDIR}/data;cd ..
	for dir in $(DIRS); do make -C $$dir package;done
	@echo "*** ${THISDIR}: completed package ***"

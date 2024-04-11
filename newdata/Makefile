TOP=..
BINDIR=${TOP}/dist/platosim
THISDIR=newdata
TEST_FAILURE_FILE=fail
DIRS=
# Check for mysql and set a variable
MYSQL_FOUND := $(shell command -v mysql 2> /dev/null)

# Conditional variable to prepend to commands
ifdef MYSQL_FOUND
DO_CMD=
else
DO_CMD=@echo "SQL not found, skipping target"; exit 0;
endif


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
	@rm -rf *.out fail config.json5
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "*** ${THISDIR}: completed clean ***"

test: config sql
	@touch $(TEST_FAILURE_FILE)
	@go test github.com/stmansour/psim/${THISDIR} && rm -f $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

sql:
	${DO_CMD}cd data; gunzip platodb.sql.gz; mysql plato <platodb.sql ; gzip platodb.sql

config:
	@cp ${TOP}/util/config.json5 .

coverage:
	go tool cover -html=coverage.out
	for dir in $(DIRS); do make -C $$dir coverage;done

package:
	mkdir -p ${BINDIR}/data
	cd data;cp platodb.csv misubclasses.csv metricssources.csv locales.csv ../${BINDIR}/data;cd ..
	for dir in $(DIRS); do make -C $$dir package;done
	@echo "*** ${THISDIR}: completed package ***"

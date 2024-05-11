TOP="../.."
DIST=${TOP}/dist/plato
THISDIR=validator
TEST_FAILURE_FILE=fail

# Check for mysql and set a variable
MYSQL_FOUND := $(shell command -v mysql 2> /dev/null)

# Conditional variable to prepend to commands
ifdef MYSQL_FOUND
DO_CMD=
else
DO_CMD=@echo "SQL not found, skipping target"; exit 0;
endif

schema: *.go
	go vet
	@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	staticcheck
	go build
	@echo "*** completed in ${THISDIR} ***"

clean:
	rm -rf data fail config.json5 "${THISDIR}"
	@echo "*** CLEAN completed in ${THISDIR} ***"

config:
	cp ${TOP}/util/config.json5 .

test:
	@echo "*** TEST completed in ${THISDIR} ***"


dovalidation: config csvdb
	./validator
	@echo "*** TEST completed in ${THISDIR} ***"

csvdb:
	mkdir -p data
	cp ${TOP}/newdata/data/platodb.csv data/
	cp ${TOP}/newdata/data/misubclasses.csv data/
	cp ${TOP}/newdata/data/metricssources.csv data/
	cp ${TOP}/newdata/data/msm.csv data/

package:
	@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	@echo "*** SECURE completed in ${THISDIR} ***"


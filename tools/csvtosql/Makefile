TOP="../.."
DIST=${TOP}/dist/platosim
THISDIR=csvtosql
TEST_FAILURE_FILE=fail

# Use environment variable to control execution
MYSQL_AVAILABLE ?= 0

# Conditional variable to prepend to commands
ifeq ($(MYSQL_AVAILABLE),1)
DO_CMD=
else
DO_CMD=@echo "MySQL not available, skipping target"; exit 0;
endif

schema: *.go
	$(DO_CMD)go vet
	$(DO_CMD)@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	$(DO_CMD)staticcheck
	$(DO_CMD)go build
	$(DO_CMD)@echo "*** completed in ${THISDIR} ***"

clean:
	rm -rf data fail config.json5 "${THISDIR}"
	@echo "*** CLEAN completed in ${THISDIR} ***"

config:
	$(DO_CMD)cp ${TOP}/util/config.json5 .

test: config csvdb
	$(DO_CMD)./csvtosql
	$(DO_CMD)@echo "*** TEST completed in ${THISDIR} ***"

db: csvdb

csvdb:
	$(DO_CMD)mkdir -p data
	$(DO_CMD)cp ${TOP}/newdata/data/platodb.csv data/
	$(DO_CMD)cp ${TOP}/newdata/data/misubclasses.csv data/
	$(DO_CMD)cp ${TOP}/newdata/data/metricssources.csv data/
	$(DO_CMD)cp ${TOP}/newdata/data/msm.csv data/

package:
	$(DO_CMD)@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	$(DO_CMD)@echo "*** SECURE completed in ${THISDIR} ***"


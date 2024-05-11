TOP="../.."
DIST=${TOP}/dist/plato
THISDIR=sqltocsv
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
	$(DO_CMD)cp ${TOP}/util/config.json5 .

test:
	$(DO_CMD)@echo "*** TEST completed in ${THISDIR} ***"

createcsv: config
	rm -rf data
	$(DO_CMD)./sqltocsv
	$(DO_CMD)@echo "*** CREATECSV completed in ${THISDIR} ***"

package:
	cp sqltocsv ${DIST}/
	$(DO_CMD)@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	$(DO_CMD)@echo "*** SECURE completed in ${THISDIR} ***"


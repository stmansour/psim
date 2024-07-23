TOP="../.."
DIST=${TOP}/dist/plato
THISDIR=tsf
TEST_FAILURE_FILE=fail

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

test: config
	@echo "*** TEST completed in ${THISDIR} ***"

package:
	mkdir -p ${DIST}/bin
	cp tsf ${DIST}/bin/
	@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	@echo "*** SECURE completed in ${THISDIR} ***"


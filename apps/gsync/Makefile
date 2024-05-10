TOP="../.."
DIST=${TOP}/dist/platosim
THISDIR=gsync
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
	cp ~/.vault/extres.json5 .

test: config
	@echo "*** TEST completed in ${THISDIR} ***"

package:
	cp gsync gfetch.sh ${DIST}/
	@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	@echo "*** SECURE completed in ${THISDIR} ***"


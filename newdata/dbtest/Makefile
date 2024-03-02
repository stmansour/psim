TOP="../.."
DIST=${TOP}/dist/platosim
THISDIR=dbtest
TEST_FAILURE_FILE=fail

schema: *.go
	go vet
	@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	staticcheck
	go build
	@echo "*** completed in ${THISDIR} ***"

clean:
	rm -f fail extres.json5 config.json5 dbtest
	@echo "*** CLEAN completed in ${THISDIR} ***"

config:
	cp ${TOP}/util/extres.json5 .
	cp ${TOP}/util/config.json5 .
	
test: config
	./dbtest
	@echo "*** TEST completed in ${THISDIR} ***"

package:
	@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	@echo "*** SECURE completed in ${THISDIR} ***"

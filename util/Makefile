TOP=..
THISDIR="util"

util: *.go
	@touch fail
	go vet
	if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	go install
	@rm -f fail
	@echo "*** Completed in ${THISDIR} ***"

clean:
	@echo "*** CLEAN completed in ${THISDIR} ***"

install:
	go install

test:
	@touch fail
	@echo "*** TEST completed in ${THISDIR} ***"
	@rm -f fail

package:
	mkdir -p ${TOP}/dist
	cp config.json5 ${TOP}/dist/
	@echo "*** PACKAGE completed in ${THISDIR} ***"

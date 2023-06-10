TOP=..
THISDIR="util"

util: *.go
	@touch fail
	go vet
	if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	go install
	@rm -f fail
	@echo "*** ${THISDIR}: completed compile ***"

clean:
	@echo "*** ${THISDIR}: completed clean ***"

install:
	go install

test:
	@touch fail
	@echo "*** ${THISDIR}: completed test ***"
	@rm -f fail

package:
	mkdir -p ${TOP}/dist
	cp config.json5 ${TOP}/dist/
	@echo "*** ${THISDIR}: completed package ***"

TOP=..
BINDIR=${TOP}/dist
THISDIR="core"

core: *.go
	go vet
	@if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	go install
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@echo "*** ${THISDIR}: completed test ***"

package:
	mkdir -p ${BINDIR}/data
	cp *.csv ${BINDIR}/data
	@echo "*** ${THISDIR}: completed package ***"

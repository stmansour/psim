TOP=../..
BINDIR=${TOP}/dist
THISDIR="apps/dr"

dr: *.go
	go vet
	@if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@echo "*** ${THISDIR}: completed test ***"

package:
	mkdir -p ${BINDIR}/
	cp er ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

TOP=../..
BINDIR=${TOP}/dist
THISDIR="apps/simulator"

dr: *.go
	go vet
	@if [[ -f "${GOPATH}/bin/golint" ]]; then golint; else staticcheck; fi
	go build
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -f simulator
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@echo "*** ${THISDIR}: completed test ***"

package:
	mkdir -p ${BINDIR}/
	cp simulator ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

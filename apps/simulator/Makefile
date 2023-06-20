TOP=../..
BINDIR=${TOP}/dist/platosim
THISDIR="apps/simulator"

dr: *.go
	go vet
	golint
	staticcheck
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

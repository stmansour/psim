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
	@go test -coverprofile=coverage.out || (echo "*** ERROR ***  go test failed!" && exit 1)
	@echo "*** ${THISDIR}: completed test ***"

coverage:
	go tool cover -html=coverage.out

package:
	mkdir -p ${BINDIR}/data
	cp *.csv ${BINDIR}/data
	@echo "*** ${THISDIR}: completed package ***"

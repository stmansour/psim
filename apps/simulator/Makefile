TOP=../..
BINDIR=${TOP}/dist/platosim
THISDIR="apps/simulator"
TEST_FAILURE_FILE = .tests_failed

dr: *.go
	go vet
	golint
	staticcheck
	go build
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -rf simulator data c*.json5 config.bak SimStats.csv coverage.out IList*.csv investorProfile.txt
	@echo "*** ${THISDIR}: completed clean ***"

test: config.json5 data/platodb.csv
	@touch $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out -coverpkg=../../... && rm -f $(TEST_FAILURE_FILE)
	@echo "*** ${THISDIR}: completed test ***"

itest: config.json5 data/platodb.csv
	./genconfigs.sh

coverage:
	go tool cover -html=coverage.out 

config.json5:
	cp ${TOP}/util/config.json5 .

data/platodb.csv:
	mkdir -p data;cp ${TOP}/data/data/platodb.csv data/

package:
	mkdir -p ${BINDIR}/
	cp simulator ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

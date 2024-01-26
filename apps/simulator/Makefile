TOP=../..
BINDIR=${TOP}/dist/platosim
THISDIR="apps/simulator"
TEST_FAILURE_FILE=fail

dr: *.go
	go vet
	golint
	staticcheck
	go build
	@echo "*** ${THISDIR}: completed ${THISDIR} ***"

clean:
	go clean
	rm -rf fail t[0-9] simulator data c*.json5 config.bak SimStats.csv coverage.out IList*.csv investorProfile.txt *Influencer.csv
	@echo "*** ${THISDIR}: completed clean ***"

test: config db
	@touch $(TEST_FAILURE_FILE)
	./functest.sh
	@go test -coverprofile=coverage.out -coverpkg=../../... && rm -f $(TEST_FAILURE_FILE)
	@echo "*** ${THISDIR}: completed test ***"

itest: config.json5 data/platodb.csv
	./genconfigs.sh

coverage:
	go tool cover -html=coverage.out 

config:
	cp ${TOP}/util/config.json5 .

db:
	mkdir -p data;cp ${TOP}/data/data/platodb.csv data/

package:
	mkdir -p ${BINDIR}/
	cp simulator ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

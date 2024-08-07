TOP=../..
BINDIR=${TOP}/dist/plato
THISDIR="apps/simulator"
TEST_FAILURE_FILE=fail
THISDIR := $(notdir $(PWD))
BUILD_TIME := $(shell date "+%Y%m%dT%H%M%S")

sim: *.go
	go vet
	golint
	staticcheck
	go build -ldflags "-X 'github.com/stmansour/psim/util.buildID=$(BUILD_TIME)'" -o simulator
	cp simwrapper.sh simwrapper ; chmod +x simwrapper
	@echo "*** $(THISDIR): completed $(THISDIR) ***"

clean: cleancsv
	go clean
	rm -rf *.prof dnalog.xlsx plato_sim*.db trace-*.csv d1.csv *.out fail [a-z][0-9] [a-z][0-9].normalized simulator data config.json5 config.bak coverage.out IList*.csv investorProfile.txt simwrapper
	@echo "*** ${THISDIR}: completed clean ***"

cleancsv:
	rm -rf crep*.csv simstats*.csv finrep*.csv invrep.csv IList*.csv arch/* dbg*.csv fitnessScores.csv c1.csv

test: config db
	@touch $(TEST_FAILURE_FILE)
	@go test -coverprofile=coverage.out -coverpkg=../../...
	./functest.sh && rm -f ${TEST_FAILURE_FILE}
	@echo "*** ${THISDIR}: completed test ***"

itest: config.json5 data/platodb.csv
	./genconfigs.sh

coverage:
	go tool cover -html=coverage.out 

config:
	cp ${TOP}/util/config.json5 .

db:
	mkdir -p data
	cp ${TOP}/newdata/data/platodb.csv data/
	cp ${TOP}/newdata/data/misubclasses.csv data/
	cp ${TOP}/newdata/data/metricssources.csv data/
	cp ${TOP}/newdata/data/msm.csv data/

package:
	mkdir -p ${BINDIR}/man/man1
	mkdir -p ${BINDIR}/bin
	cp simulator simwrapper ${BINDIR}/bin/
	cp simulator.1 ${BINDIR}/man/man1
	@echo "*** ${THISDIR}: completed package ***"

releaseman:
	if [ -d /usr/local/share/man/man1 ] && [ -w /usr/local/share/man/man1 ]; then cp simulator.1 /usr/local/share/man/man1/ ; fi
	@echo "*** ${THISDIR}: completed releaseman ***"

refmt:
	fmt simulator.1 > simulator.11 ; mv simulator.11 simulator.1

release: simulator
	cp simulator /usr/local/plato/bin/
	cp simwrapper /usr/local/plato/bin/


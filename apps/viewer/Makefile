TOP=../..
BINDIR=${TOP}/dist/platosim
THISDIR="apps/iewer"
TEST_FAILURE_FILE=fail

clean:
	rm -rf data *.out fail config.json5
	@echo "*** ${THISDIR}: completed clean ***"

test:
	@echo "*** ${THISDIR}: completed package ***"

db:
	mkdir -p data;cp ${TOP}/newdata/data/platodb.csv data/

package:
	mkdir -p ${BINDIR}
	cp index.html ${BINDIR}/
	cp app.js ${BINDIR}/
	@echo "*** ${THISDIR}: completed package ***"

release:
	cp index.html /Library/Webserver/Documents/viewer/
	cp app.js /Library/Webserver/Documents/viewer/
	cp -r data /Library/Webserver/Documents/viewer/

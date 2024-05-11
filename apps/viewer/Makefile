OS := $(shell uname)

TOP=../..
BINDIR=${TOP}/dist/plato
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

release: db
ifeq ($(OS),Linux)
	@echo "Running Linux-specific commands"
	cp index.html /var/www/html/viewer/
	cp app.js /var/www/html/viewer/
	cp -r data /var/www/html/viewer/

else ifeq ($(OS),Darwin)
	@echo "Running MacOS-specific commands"
	cp index.html /Library/Webserver/Documents/viewer/
	cp app.js /Library/Webserver/Documents/viewer/
	cp -r data /Library/Webserver/Documents/viewer/
endif

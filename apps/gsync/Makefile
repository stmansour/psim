TOP="../.."
DIST=${TOP}/dist/plato
THISDIR=gsync
TEST_FAILURE_FILE=fail

gsync: *.go gfetch.sh
	go vet
	@golint ./... | grep -v "use underscores in Go names;" | ( ! grep . )
	staticcheck
	go build
	cp gfetch.sh gfetch
	chmod +x gfetch
	@echo "*** completed in ${THISDIR} ***"

clean:
	rm -rf extres.json5 masterlist.txt gfetch.log urls.txt data fail config.json5 "${THISDIR}"
	@echo "*** CLEAN completed in ${THISDIR} ***"

config:
	cp ${TOP}/util/config.json5 .
	cp ~/.vault/extres.json5 .

test: config
	@echo "*** TEST completed in ${THISDIR} ***"

package:
	mkdir -p ${DIST}/man/man1
	cp gfetch.1 ${DIST}/man/man1
	cp gsync gfetch ${DIST}/
	@echo "*** PACKAGE completed in ${THISDIR} ***"

secure:
	@echo "*** SECURE completed in ${THISDIR} ***"

refmt:
	fmt howto.txt > howto.txt1 ; mv howto.txt1 howto.txt
	fmt gfetch.1 > gfetch.11 ; mv gfetch.11 gfetch.1

releaseman:
	if [ -d /usr/local/share/man/man1 ] && [ -w /usr/local/share/man/man1 ]; then cp gfetch.1 /usr/local/share/man/man1/ ; fi

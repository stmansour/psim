DIRS=csvtosql validator sqltocsv
THISDIR=tools

# Conditional variable to prepend to commands
ifdef MYSQL_AVAILABLE
DO_CMD=
else
DO_CMD=@echo "SQL not found, skipping target"; exit 0;
endif



apps:
	${DO_CMD}for dir in $(DIRS); do make -C $$dir;done
	@echo "*** ${THISDIR}: completed compile ***"

clean:
	${DO_CMD}for dir in $(DIRS); do make -C $$dir clean;done
	@echo "*** ${THISDIR}: completed clean ***"

test:
	${DO_CMD}for dir in $(DIRS); do make -C $$dir test;done
	@echo "*** ${THISDIR}: completed test ***"

package:
	${DO_CMD}for dir in $(DIRS); do make -C $$dir package;done
	@echo "*** ${THISDIR}: completed package ***"


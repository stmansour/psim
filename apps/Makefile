DIRS=simulator simtalk viewer
THISDIR=apps

apps:
	for dir in $(DIRS); do make -C $$dir;done
	@echo "*** ${THISDIR}: completed compile ***"

clean:
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "*** ${THISDIR}: completed clean ***"

test:
	for dir in $(DIRS); do make -C $$dir test;done
	@echo "*** ${THISDIR}: completed test ***"

package:
	for dir in $(DIRS); do make -C $$dir package;done
	@echo "*** ${THISDIR}: completed package ***"


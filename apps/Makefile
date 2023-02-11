DIRS=dr er

apps:
	for dir in $(DIRS); do make -C $$dir;done

clean:
	for dir in $(DIRS); do make -C $$dir clean;done

test:
	for dir in $(DIRS); do make -C $$dir test;done

package:
	for dir in $(DIRS); do make -C $$dir package;done


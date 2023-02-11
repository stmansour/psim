DIRS=data apps
DIST=dist

.PHONY: test

psim:
	for dir in $(DIRS); do make -C $$dir;done

clean:
	rm -rf dist
	for dir in $(DIRS); do make -C $$dir clean;done

test:
	for dir in $(DIRS); do make -C $$dir test;done

package:
	for dir in $(DIRS); do make -C $$dir package;done

all: clean psim package test
	echo "Completed"

build: clean psim package

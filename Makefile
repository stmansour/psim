# test comment
DIRS=util data apps
DIST=dist
TEST_FAILURE_FILE = .tests_failed

.PHONY: test

psim:
	for dir in $(DIRS); do make -C $$dir;done

clean:
	rm -rf dist
	for dir in $(DIRS); do make -C $$dir clean;done

test: do_tests check_tests check_coverage

do_tests:
	@echo "******************************************************************"
	@echo "                          TESTS"
	@echo "******************************************************************"
	for dir in $(DIRS); do make -C $$dir test;done

check_tests:
	@echo "Checking tests..."
	@if test -n "$(shell find . -name .tests_failed)"; then \
		echo "Tests have failed in the following directories:"; \
		find . -name .tests_failed -exec dirname {} \; ; \
		exit 1; \
	else \
		echo "All tests passed."; \
	fi

check_coverage:
	@echo "Checking coverage..."
	@for dir in $(shell find . -name coverage.out); do \
		coverage=$$(go tool cover -func=$$dir | grep total | awk '{print $$NF}') ; \
		echo "`dirname $$dir` : $$coverage"; \
	done

package:
	for dir in $(DIRS); do make -C $$dir package;done

all: clean psim package test
	echo "Completed"

build: clean psim package

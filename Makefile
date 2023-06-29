# test comment
DIRS=util data core tools apps
DIST=dist 
TEST_FAILURE_FILE = .tests_failed

.PHONY: test

psim:
	for dir in $(DIRS); do make -C $$dir;done

clean:
	rm -rf dist
	for dir in $(DIRS); do make -C $$dir clean;done

test: do_tests check_tests

do_tests:
	@echo "------------------------------------------------------------------"
	@echo "                          TESTS"
	@echo "------------------------------------------------------------------"
	for dir in $(DIRS); do make -C $$dir test;done

check_tests:
	@echo "------------------------------------------------------------------"
	@echo "                      TESTS RESULTS"
	@echo "------------------------------------------------------------------"
	@echo
	@echo "UNIT TEST CODE COVERAGE"
	@echo "======================="
	@for dir in $(shell find . -name coverage.out); do \
		if [ "$$dir" != "./apps/simulator/coverage.out" ]; then \
		coverage=$$(go tool cover -func=$$dir | grep total | awk '{print $$NF}') ; \
		echo "`dirname $$dir` : $$coverage"; \
		fi \
	done
	@echo
	@echo "FUNCTIONAL TEST CODE COVERAGE"
	@echo "============================="
	@for dir in $(shell find ./apps -name coverage.out); do \
		coverage=$$(go tool cover -func=$$dir | grep total | awk '{print $$NF}') ; \
		echo "`dirname $$dir` : $$coverage"; \
	done
	@echo
	@if test -n "$(shell find . -name .tests_failed)"; then \
		echo "Tests have failed in the following directories:"; \
		find . -name .tests_failed -exec dirname {} \; ; \
			exit 1; \
		else \
			echo "****************************"; \
			echo "*     ALL TESTS PASSED     *"; \
			echo "****************************"; \
		fi

package:
	for dir in $(DIRS); do make -C $$dir package;done
	cd dist ; tar cvf platosim.tar platosim ; gzip platosim.tar

all: clean psim package test
	@echo "Completed"

build: clean psim package

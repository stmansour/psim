# test comment
DIRS=util sqlt newdata newcore tools apps
DIST=dist 
TEST_FAILURE_FILE=fail
# Temporary file for storing start time
TIMER_FILE := .build_timer

.PHONY: install-tools golint staticcheck test

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
	@if test -n "$(shell find . -name ${TEST_FAILURE_FILE})"; then \
		echo "Tests have failed in the following directories:"; \
		find . -name "${TEST_FAILURE_FILE}" -exec dirname {} \; ; \
			exit 1; \
		else \
			echo "****************************"; \
			echo "***   ALL TESTS PASSED   ***"; \
			echo "****************************"; \
		fi

package:
	for dir in $(DIRS); do make -C $$dir package;done
	if [ -d /usr/local/plato/bin/data ]; then cd dist/plato/bin ; rm -rf data ; cp -r /usr/local/plato/bin/data . ; fi
	./mkdist.sh -c

post:
	./mkdist.sh -p

all: starttimer clean psim package test stoptimer
	@echo "Completed"

build: starttimer clean psim package stoptimer

stats:
	@find . -name "*.go" | srcstats

release:
	cp -r dist/plato /usr/local/plato/;cd /usr/local/plato;rm -rf bin;mv plato bin;
	if [ -d /usr/local/share/man/man1 ] && [ -w /usr/local/share/man/man1 ]; then cp ./dist/plato/man/man1/* /usr/local/share/man/man1/ ; fi
	@if [ -f ~/.vault/extres.json5 ]; then \
	    cp ~/.vault/extres.json5 /usr/local/plato/bin/ ; chmod 755 /usr/local/plato/bin/extres.json5 ;\
	else \
	    echo "extres.json5 was not found, some apps will not function without it." ; \
	fi

refmt:
	fmt design.txt > design.txt1 ; mv design.txt1 design.txt

install-tools:
	go get -u github.com/go-sql-driver/mysql
	go get github.com/google/uuid
	go install golang.org/x/lint/golint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

starttimer:
	@echo $$(date +%s) > $(TIMER_FILE)

stoptimer:
	@start=$$(cat $(TIMER_FILE)); \
	end=$$(date +%s); \
	elapsed=$$((end - start)); \
	hours=$$((elapsed / 3600)); \
	minutes=$$(( (elapsed / 60) % 60 )); \
	seconds=$$((elapsed % 60)); \
	if [ $$hours -gt 0 ]; then \
		echo "Elapsed time: $$hours hour(s) $$minutes minute(s) $$seconds second(s)"; \
	elif [ $$minutes -gt 0 ]; then \
		echo "Elapsed time: $$minutes minute(s) $$seconds second(s)"; \
	else \
		echo "Elapsed time: $$seconds second(s)"; \
	fi; \
	rm -f $(TIMER_FILE)


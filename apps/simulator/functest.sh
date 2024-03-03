#!/bin/bash

RUNSINGLETEST=0
TESTNAME="TestSimulator"
TESTCOUNT=0
ARCHIVE=arch

usage() {
    cat <<EOF

SYNOPSIS
	$0 [-a -t]

	Run the tests and compare the output of each test step to its associated
    known-good output. If they miscompare, fail and stop the script. If they
    match, keep going until all tasks are completed.

OPTIONS
	-a  If a test fails, pause after showing diffs from gold files, prompt
	    for what to do next:  [Enter] to continue, m to move the output file
	    into gold/ , or Q / X to exit.

    -h  display this help information.

	-t  Sets the environment variable RUNSINGLETEST to the supplied value. By
	    default, "${RUNSINGLETEST}x" == "x" and this should cause all of the
	    tests in the script to run. But if you would like to be able to run
	    an individual test by name, you can use ${RUNSINGLETEST} to check and
	    see if the user has requested a specific test.
EOF
}

passmsg() {
    t="${TESTNAME}"
    printf "PASSED  %-20.20s  %-40.40s \n" "${TESTDIR}" "${t}" ${TESTCOUNT}
}

failmsg() {
    t="${TESTNAME}"
    printf "FAILED  %-20.20s  %-40.40s \n" "${TESTDIR}" "${t}" ${TESTCOUNT}
}

#------------------------------------------------------------------------------
# Function to compare a report file to its gold standard
# INPUTS
#    $1 = name of un-normalized output file
#------------------------------------------------------------------------------
compareToGold() {
    local reportFile=$1
    local goldFile="gold/${reportFile}.gold"
    local normalizedFile="${reportFile}.normalized"

    # if it's a csv file, delete to the first blank line...
    if [[ ${reportFile} =~ \.csv$ ]]; then
        awk 'flag; /^$/ {flag=1}' "${reportFile}" >"${reportFile}.tmp" && mv "${reportFile}.tmp" "${reportFile}"
    fi

    # Normalize the report file
    sed -E \
        -e 's/Version:[[:space:]]+[0-9]+\.[0-9]+-[0-9]{8}-[0-9]{6}/Version: VERSION_PLACEHOLDER/' \
        -e 's/Random number seed:[[:space:]]+[0-9]+/Random number seed: SEED_PLACEHOLDER/' \
        -e 's/Archive directory:.*/Archive directory: PLACEHOLDER/' \
        "$reportFile" >"$normalizedFile"

    # Check if running on Windows
    if [[ "$(uname -s)" =~ MINGW|CYGWIN|MSYS ]]; then
        echo "Detected Windows OS. Normalizing line endings for ${normalizedFile}."

        # Use sed to replace CRLF with LF, output to temp file
        sed 's/\r$//' "${normalizedFile}" >"${goldFile}.tmp"
        goldFile="${goldFile}.tmp"
    fi

    # Compare the normalized report to the gold standard
    if diff "${normalizedFile}" "${goldFile}"; then
        echo "PASSED"
        rm "${normalizedFile}"
    else
        echo "Differences detected."
        # Prompt the user for action
        if [[ "${ASKBEFOREEXIT}" == 1 ]]; then
            while true; do
                read -rp "Choose action - Continue (C), Move (M), or eXit (X) [C]: " choice
                choice=${choice:-C} # Default to 'C' if no input
                case "$choice" in
                C | "")
                    echo "Continuing..."
                    return 0
                    ;;
                M | m)
                    echo "Moving normalized file to gold standard..."
                    mv "$normalizedFile" "$goldFile"
                    return 0
                    ;;
                X | x)
                    echo "Exiting..."
                    exit 1
                    ;;
                *) echo "Invalid choice. Please choose C, M, or X." ;;
                esac
            done
        fi
    fi
}

###############################################################################
#    INPUT
###############################################################################
while getopts "at:" o; do
    echo "o = ${o}"
    case "${o}" in
    a)
        ASKBEFOREEXIT=1
        echo "WILL ASK BEFORE EXITING ON ERROR"
        ;;
    t)
        SINGLETEST="${OPTARG}"
        echo "SINGLETEST set to ${SINGLETEST}"
        ;;
    *)
        usage
        exit 1
        ;;
    esac
done
shift $((OPTIND - 1))
############################################################################################

if [ ! -d data ]; then
    echo "there is no data/ directory"
    echo "please run 'make db' or create data/ and put a csv database in it"
    exit 1
fi
if [ ! -f data/platodb.csv ]; then
    echo "there is no database in data/"
    echo "please run 'make db' or put a csv database in data/"
    exit 1
fi

mkdir -p "${ARCHIVE}"

#------------------------------------------------------------------------------
#  TEST a
#  ping the server
#
#  Scenario:
#  Execute the url to ping the server
#
#  Expected Results:
#   1.  It should return the server version
#------------------------------------------------------------------------------
TFILES="a"
STEP=0
if [[ "${SINGLETEST}${TFILES}" = "${TFILES}" || "${SINGLETEST}${TFILES}" = "${TFILES}${TFILES}" ]]; then
    echo -n "Test ${TFILES} - "
    echo -n "Single Investor test... "
    RESFILE="${TFILES}${STEP}"
    ./simulator -ar -adir "${ARCHIVE}" -trace -c singleInvestor.json5 >"${RESFILE}"
    compareToGold ${RESFILE}
    ((TESTCOUNT++))
fi

TFILES="b"
STEP=0
if [[ "${SINGLETEST}${TFILES}" = "${TFILES}" || "${SINGLETEST}${TFILES}" = "${TFILES}${TFILES}" ]]; then
    echo -n "Test ${TFILES} - "
    echo -n "Linguistic Influencers test... "
    RESFILE="${TFILES}${STEP}"
    ./simulator -ar -adir "${ARCHIVE}" -trace -c linguistics.json5 >"${RESFILE}"
    compareToGold ${RESFILE}
    ((TESTCOUNT++))
fi

TFILES="c"
STEP=0
if [[ "${SINGLETEST}${TFILES}" = "${TFILES}" || "${SINGLETEST}${TFILES}" = "${TFILES}${TFILES}" ]]; then
    echo -n "Test ${TFILES} - "
    echo -n "Crucible test..."
    RESFILE="${TFILES}${STEP}"
    ./simulator -C -c confcru.json5 >"${RESFILE}"
    compareToGold ${RESFILE}
    mv crep.csv c1.csv
    compareToGold c1.csv
    ((TESTCOUNT++))
fi

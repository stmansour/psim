#!/bin/bash

RUNSINGLETEST=0
TESTCOUNT=0
ERRORCOUNT=0
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
        -e 's/^Version:.*/Version: VERSION_PLACEHOLDER/' \
        -e 's/^Available cores:.*/Version: PLACEHOLDER/' \
        -e 's/Random number seed:[[:space:]]+[0-9]+/Random number seed: SEED_PLACEHOLDER/' \
        -e 's/Archive directory:.*/Archive directory: PLACEHOLDER/' \
        -e 's/Elapsed time:.*/Archive directory: PLACEHOLDER/' \
        -e 's/Investor-[0-9a-zA-Z-]+/Investor-GUID/' \
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
        echo "Differences detected.  meld ${normalizedFile} ${goldFile}"
        ((ERRORCOUNT++))
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
#  single investor test with trace
#------------------------------------------------------------------------------
TFILES="a"
STEP=0
if [[ "${SINGLETEST}${TFILES}" = "${TFILES}" || "${SINGLETEST}${TFILES}" = "${TFILES}${TFILES}" ]]; then
    echo -n "Test ${TFILES} - "
    echo -n "Single Investor test... "
    RESFILE="${TFILES}${STEP}"
    ./simulator -ar -adir "${ARCHIVE}" -notalk -trace -c singleInvestor.json5 >"${RESFILE}"
    compareToGold ${RESFILE}
    ((TESTCOUNT++))
fi

#------------------------------------------------------------------------------
#  TEST b
#  test linguistic influencer
#------------------------------------------------------------------------------
TFILES="b"
STEP=0
if [[ "${SINGLETEST}${TFILES}" = "${TFILES}" || "${SINGLETEST}${TFILES}" = "${TFILES}${TFILES}" ]]; then
    echo -n "Test ${TFILES} - "
    echo -n "Linguistic Influencers test... "
    RESFILE="${TFILES}${STEP}"
    ./simulator -ar -adir "${ARCHIVE}" -notalk -trace -c linguistics.json5 >"${RESFILE}"
    compareToGold ${RESFILE}
    ((TESTCOUNT++))
fi

#------------------------------------------------------------------------------
#  TEST c
#  test crucible
#------------------------------------------------------------------------------
TFILES="c"
STEP=0
if [[ "${SINGLETEST}${TFILES}" = "${TFILES}" || "${SINGLETEST}${TFILES}" = "${TFILES}${TFILES}" ]]; then
    echo -n "Test ${TFILES} - "
    echo -n "Crucible test..."
    RESFILE="${TFILES}${STEP}"
    ./simulator -C -c confcru.json5 -notalk >"${RESFILE}"
    compareToGold ${RESFILE}
    mv crep.csv c1.csv
    compareToGold c1.csv
    ((TESTCOUNT++))
fi


echo "Total tests: ${TESTCOUNT}"
echo "Total errors: ${ERRORCOUNT}"
if [ "${ERRORCOUNT}" -gt 0 ]; then
    exit 2
fi

#!/usr/bin/env bash

pkill -f Peerster

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="false"

FAILED=0
SUCCESS=0

expect_contains() {

  file="logs/${1}.out"
  regex=${2}

  if (grep -q "$regex" ${file}) ; then
    SUCCESS=$((SUCCESS+1))
    echo -e "${GREEN}- ${file} : <CONTAINS> ${regex}${NC}"
  else
    FAILED=$((FAILED+1))
    echo -e "${RED}- ${file} : <SHOULD CONTAIN> ${regex}${NC}"
  fi
}

expect_missing() {

  file="logs/${1}.out"
  regex=${2}

  if !(grep -q "$regex" ${file}) ; then
    SUCCESS=$((SUCCESS+1))
    echo -e "${GREEN}- ${file} : <DOES NOT CONTAIN> ${regex}${NC}"
  else
    FAILED=$((FAILED+1))
    echo -e "${RED}- ${file} : <SHOULD NOT CONTAIN> ${regex}${NC}"
  fi
}

print_test_results() {

  echo ""
  if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}  TESTS SUCCEEDED (${SUCCESS} tests total)${NC}"
  else
    echo -e "${RED}  TESTS FAILED WITH ${FAILED} FAILURES AND ${SUCCESS} SUCCESSES${NC}"
  fi
  echo ""
}

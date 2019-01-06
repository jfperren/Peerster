#!/usr/bin/env bash

# Build

CRYPTOOPTS=""
DEBUG=false
PACKAGE=false
while [[ $# -gt 0 ]]
do
    key="$1"

    case $key in
        -v|--verbose|-d|--debug)
            DEBUG=true
            ;;
        --package)
            PACKAGE=true
            ;;
        -c|--crypto)
            shift
            if [[ "$1" == 1 ]]
            then
                CRYPTOOPTS=" -sign-only"
            elif [[ "$1" == 2 ]]
            then
                CRYPTOOPTS=" -cypher-if-possible"
            fi
            ;;
        *)
            # unknown option
            ;;
    esac
    shift
done

if ! $PACKAGE; then
	source ./scripts/build.sh
	source ./tests/sh/helpers.sh
fi

# Variables

outputFiles=()

UIPort=8080
gossipPort=5000
name="A"
rtimer=1

message_a_c="This is a message from A to C; should be delivered"
# message_a_d="This is a message from A to D; should not be delivered"
# message_j_a="This is a message from J to A; should be delivered"
# message_g_c="This is a message from G to C; should be delivered"
# message_j_b="This is a message from J to B; should not be delivered"

# Create Gossipers

for i in `seq 1 4`;
do
	outFileName="logs/$name.out"

  peerPort=$((($i)%4+5000))

	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
  sharedDirName="$sharedDir/$name/"
  downloadDirName="$downloadDir/$name/"
  rtimer=1

  if [[ $name == "D" ]]; then
    mode=""
  else
    mode="-sign-only"
  fi

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer -verbose $mode 2> $outFileName &

	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi

  outputFiles+=("$outFileName")
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

# Send Private Messages

sleep 30

echo -e "${NC}# CHECK that A, B, C skip when not authenticated${NC}"

expect_contains A "NOT AUTHENTICATED skip send packet"
expect_contains B "NOT AUTHENTICATED skip send packet"
expect_contains C "NOT AUTHENTICATED skip send packet"

echo -e "${NC}# CHECK that D does not need to authenticate${NC}"

expect_missing D "NOT AUTHENTICATED skip send packet"

echo -e "${NC}# CHECK that chain converged${NC}"

hex="[a-zA-Z0-9 :,.]*"

expect_contains A "CHAIN ${hex}[ABC]${hex}[ABC]${hex}[ABC]"
expect_contains B "CHAIN ${hex}[ABC]${hex}[ABC]${hex}[ABC]"
expect_contains C "CHAIN ${hex}[ABC]${hex}[ABC]${hex}[ABC]"
expect_contains D "CHAIN ${hex}[ABC]${hex}[ABC]${hex}[ABC]"

echo -e "${NC}# CHECK that D is not in chain${NC}"

expect_missing A "CHAIN ${hex}D"
expect_missing B "CHAIN ${hex}D"
expect_missing C "CHAIN ${hex}D"
expect_missing D "CHAIN ${hex}D"

./client/client -UIPort=8080 -msg="$message_a_c" -dest="C"
# ./client/client -UIPort=8080 -msg="$message_a_d" -dest="D"
# ./client/client -UIPort=8089 -msg="$message_j_a" -dest="A"
# ./client/client -UIPort=8086 -msg="$message_g_c" -dest="C"
# ./client/client -UIPort=8089 -msg="$message_j_b" -dest="B"

sleep 20

pkill -f Peerster

# Tests

echo -e "${NC}# CHECK that A knew where to send${NC}"

expect_missing A "UNKNOWN DESTINATION C"

echo -e "${NC}# CHECK that message was sent and forwarded${NC}"

expect_contains A "ROUTE POINT-TO-POINT MESSAGE destination C"
expect_contains B "ROUTE POINT-TO-POINT MESSAGE destination C"

echo -e "${NC}# CHECK that valid recipients got their message${NC}"

expect_contains C "PRIVATE origin A hop-limit "
expect_contains C " contents $message_a_c"

if [[ $PACKAGE == false ]]; then
	print_test_results
fi

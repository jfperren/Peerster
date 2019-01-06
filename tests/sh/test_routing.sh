#!/usr/bin/env bash

# The general setup is the following:
#
# There is a ring of 7 nodes (similar to test_2_ring but with 7):
#    A - B - C - D - E - F - G - (A)
#
# In this ring, only A, D and F send route rumors, the others are in
# "invisible mode". Therefore, only A, D and F should be visible on the network
#
# On top of the ring, 3 additional nodes H, I and J are connected to A, D, F
# respectively. Plus,
#
#  - B sends a regular rumor at some point
#  - H sends route rumors periodically
#  - I does not send route rumors but sends a regular rumor at some point
#  - J does not send route rumors and does not send regular rumors
#
# In this setup, we expect that
#
# - A, D, F and H are visible by everybody because they send route rumors
# - B and I are also visible because they send regular rumors
# - The other nodes are not visible because they don't send any rumor

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
            source ./scripts/build.sh
            source ./tests/sh/helpers.sh
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


# Preparation

outputFiles=()

UIPort=8080
gossipPort=5000
name="A"
rtimer=1

# Start Gossipers

for i in `seq 1 10`;
do
	outFileName="logs/$name.out"

  if [ $name = "H" ]; then
    peerPort=5000
  elif [ $name = "I" ]; then
    peerPort=5002
  elif [ $name = "J" ]; then
    peerPort=5004
  else
    peerPort=$((($i)%7+5000))
  fi

	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
  sharedDirName="$sharedDir/$name/"
  downloadDirName="$downloadDir/$name/"

  if [ $name = "A" ] || [ $name = "C" ] || [ $name = "F" ] || [ $name = "H" ]; then
    rtimer=1
  else
    rtimer=0
  fi

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer -verbose$CRYPTOOPTS > $outFileName &

	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi

  outputFiles+=("$outFileName")
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
    if [[ "$CRYPTOOPTS" != "" ]]
    then
        sleep 3
    fi
done

# Nothing to do except we send one rumor

sleep 2
if [[ "$CRYPTOOPTS" != "" ]]
then
    sleep 10
fi

./client/client -UIPort=8088 -msg="Hello"

sleep 2
if [[ "$CRYPTOOPTS" != "" ]]
then
    sleep 10
fi

pkill -f Peerster

# Tests

echo -e "${NC}# CHECK that visible nodes updated their routing table correctly${NC}"

expect_contains A "DSDV C"
expect_contains A "DSDV F"

expect_contains C "DSDV A"
expect_contains C "DSDV F"

expect_contains F "DSDV A"
expect_contains F "DSDV C"

echo -e "${NC}# CHECK that other nodes can see as well${NC}"

expect_contains B "DSDV A"
expect_contains B "DSDV C"
expect_contains B "DSDV F"

expect_contains D "DSDV A"
expect_contains D "DSDV C"
expect_contains D "DSDV F"

expect_contains E "DSDV A"
expect_contains E "DSDV C"
expect_contains E "DSDV F"

echo -e "${NC}# CHECK that silent nodes are not seen${NC}"

expect_missing A "DSDV B"
expect_missing A "DSDV D"
expect_missing A "DSDV E"

expect_missing C "DSDV B"
expect_missing C "DSDV D"
expect_missing C "DSDV E"

expect_missing F "DSDV B"
expect_missing F "DSDV D"
expect_missing F "DSDV E"

echo -e "${NC}# CHECK that A overwrite its routing table${NC}"

# expect_contains A "DSDV H 127.0.0.1:5001"
# expect_contains A "DSDV H 127.0.0.1:5009"

echo -e "${NC}# CHECK that H and I update their routing table${NC}"

expect_contains H "DSDV A 127.0.0.1:5000"
expect_contains H "DSDV C 127.0.0.1:5000"
expect_contains H "DSDV F 127.0.0.1:5000"

expect_contains I "DSDV A 127.0.0.1:5002"
expect_contains I "DSDV C 127.0.0.1:5002"
expect_contains I "DSDV F 127.0.0.1:5002"

echo -e "${NC}# CHECK that J also update its routing table${NC}"

expect_contains J "DSDV A 127.0.0.1:5004"
expect_contains J "DSDV C 127.0.0.1:5004"
expect_contains J "DSDV F 127.0.0.1:5004"

echo -e "${NC}# CHECK that H and I become visible${NC}"

expect_contains A "DSDV H 127.0.0.1:5007"
expect_contains A "DSDV I 127.0.0.1:5001"

expect_contains B "DSDV H 127.0.0.1:5000"
expect_contains B "DSDV I 127.0.0.1:5002"

expect_contains C "DSDV H 127.0.0.1:5001"
expect_contains C "DSDV I"

expect_contains D "DSDV H"
expect_contains D "DSDV I"

expect_contains E "DSDV H"
expect_contains E "DSDV I"

echo -e "${NC}# CHECK that J is not visible${NC}"

expect_missing A "DSDV J"
expect_missing B "DSDV J"
expect_missing C "DSDV J"
expect_missing D "DSDV J"
expect_missing E "DSDV J"


# echo -e "${NC}# Check that E received the file from A${NC}"
#
# expect_contains E "DOWNLOADING metafile of $file_a_small from A"
# expect_contains E "DOWNLOADING $file_a_small chunk 0 from A"
# expect_contains E "RECONSTRUCTED file $file_a_small"
#
# echo -e "${NC}# Check that A received the file from E${NC}"
#
# expect_contains A "DOWNLOADING metafile of $file_e_big from E"
# expect_contains A "DOWNLOADING $file_e_big chunk 0 from E"
# expect_contains A "DOWNLOADING $file_e_big chunk 50 from E"
# expect_contains A "DOWNLOADING $file_e_big chunk 100 from E"
# expect_contains A "DOWNLOADING $file_e_big chunk 150 from E"
# expect_contains A "DOWNLOADING $file_e_big chunk 179 from E"
# expect_contains A "RECONSTRUCTED file $file_e_big"
#
# echo -e "${NC}# Check that A received the file from C${NC}"
#
# expect_contains A "DOWNLOADING metafile of $file_c_medium from C"
# expect_contains A "DOWNLOADING $file_c_medium chunk 0 from C"
# expect_contains A "DOWNLOADING $file_c_medium chunk 5 from C"
# expect_contains A "DOWNLOADING $file_c_medium chunk 10 from C"
# expect_contains A "RECONSTRUCTED file $file_c_medium"
#
# echo -e "${NC}# Check that E received the file from C${NC}"
#
# expect_contains E "DOWNLOADING metafile of $file_c_medium from C"
# expect_contains E "DOWNLOADING $file_c_medium chunk 0 from C"
# expect_contains E "DOWNLOADING $file_c_medium chunk 5 from C"
# expect_contains E "DOWNLOADING $file_c_medium chunk 10 from C"
# expect_contains E "RECONSTRUCTED file $file_c_medium"
#
# echo -e "${NC}# Check that B did not find F${NC}"
#
# expect_contains B "UNKNOWN DESTINATION F"
#
# echo -e "${NC}# Check that B did not process the file that is too big${NC}"
#
# expect_contains B "WARNING file too_big.txt is too big for Peerster (max. 2Mb)"
#
# echo -e "${NC}# Check that D looked and did not find file asked by A${NC}"
#
# expect_contains D "RECEIVE DATA REQUEST from A to D metahash $hash_file_d_inexistant"
# expect_contains D "NOT FOUND hash $hash_file_d_inexistant from A"

if [[ $PACKAGE = false ]]; then
	print_test_results
fi

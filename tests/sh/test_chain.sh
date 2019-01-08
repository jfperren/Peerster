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
nb_nodes=10
while [[ $# -gt 0 ]]
do
    key="$1"

    case $key in
        -v|--verbose|-d|--debug)
            DEBUG=true
            ;;
        --package)
            source ./scripts/build.sh
            source ./tests/sh/helpers.sh
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
        -n|--nb-nodes)
            shift
            nb_nodes="$1"
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

file_a="hello.txt"
file_b="image.txt"
file_c="message.txt"

downloadDir="_Downloads"
sharedDir="_SharedFiles"

# Start Gossipers
for i in `seq 1 $nb_nodes`;
do
	outFileName="logs/$name.out"

  if [ $name = "H" ]; then
    peerPort=5000
  elif [ $name = "I" ]; then
    peerPort=5002
  elif [ $name = "J" ]; then
    peerPort=5004
  else
      if [[ $nb_nodes > 6 ]]
      then
          peerPort=$((($i)%7+5000))
      else
          peerPort=$(($i%$nb_nodes+5000))
      fi
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

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer -verbose -separatefs$CRYPTOOPTS 2> $outFileName &

	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi

    outputFiles+=("$outFileName")
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
    if [[ "$CRYPTOOPTS" != "" ]]
    then
        sleep 2
    fi
done

# Create files

touch "$sharedDir/A/$file_a"
echo "Hello" > "$sharedDir/A/$file_a"

touch "$sharedDir/E/$file_a"
echo "Hello" > "$sharedDir/E/$file_a"

touch "$sharedDir/B/$file_b"
echo "Image" > "$sharedDir/B/$file_b"

touch "$sharedDir/D/$file_b"
echo "Image" > "$sharedDir/D/$file_b"

touch "$sharedDir/E/$file_c"
echo "Message" > "$sharedDir/E/$file_c"

# Nothing to do here, we just let them mine

sleep 2
if [[ "$CRYPTOOPTS" != "" ]]
then
    sleep 5
fi

./client/client -UIPort=8080 -file=$file_a
./client/client -UIPort=8081 -file=$file_b

sleep 4
if [[ $nb_nodes > 4 ]]
then
    if [[ "$CRYPTOOPTS" != "" ]]
    then
        sleep 10
    fi

    ./client/client -UIPort=8084 -file=$file_a
    ./client/client -UIPort=8084 -file=$file_c

    sleep 2
fi
if [[ $nb_nodes > 3 ]]
then
    if [[ "$CRYPTOOPTS" != "" ]]
    then
        sleep 5
    fi

    ./client/client -UIPort=8083 -file=$file_b

    sleep 4
fi
if [[ "$CRYPTOOPTS" != "" ]]
then
    sleep 10
fi

pkill -f Peerster

# Tests

echo -e "${NC}# CHECK that everybody has found at least one block${NC}"

for i in `seq 0 $(($nb_nodes - 1))`
do
    expect_contains "${outputFiles[$i]:5:1}" "FOUND-BLOCK"
done

echo -e "${NC}# CHECK that everybody is printing the chain${NC}"

for i in `seq 0 $(($nb_nodes - 1))`
do
    expect_contains "${outputFiles[$i]:5:1}" "CHAIN"
done

echo -e "${NC}# Check that E received the file from A${NC}"

for i in `seq 0 $(($nb_nodes - 1))`
do
    expect_contains "${outputFiles[$i]:5:1}" "FORK-LONGER rewind . blocks"
done

echo -e "${NC}# Check that A, B and E successfully added its file${NC}"

expect_contains A "CANDIDATE transaction $file_a| successfully added"
expect_contains B "CANDIDATE transaction $file_b| successfully added"
if [[ $nb_nodes > 4 ]]
then
    expect_contains E "CANDIDATE transaction $file_c| successfully added"
    echo -e "${NC}# Check that D and E do not add their duplicate names${NC}"

    expect_contains E "IGNORE transaction $file_a| already in chain"
    expect_contains D "IGNORE transaction $file_b| already in chain"
fi

echo -e "${NC}# Check that A,B and E successfully transmitted its file${NC}"

expect_contains A "BROADCAST transaction $file_a|"
expect_contains B "BROADCAST transaction $file_b|"
if [[ $nb_nodes > 4 ]]
then
    expect_contains E "BROADCAST transaction $file_c|"
fi

echo -e "${NC}# Check that everybody has the files${NC}"

files="$file_a $file_b"
if [[ $nb_nodes > 4 ]]
then
    files+=" $file_c"
fi

for j in $files
do
    for i in `seq 0 $(($nb_nodes - 1))`
    do
        expect_contains "${outputFiles[$i]:5:1}" "CHAIN [a-zA-Z0-9 :,.]*$j"
    done
done

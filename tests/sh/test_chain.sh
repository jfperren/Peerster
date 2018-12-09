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

if [[ $* != *--package* ]]; then
	source ./scripts/build.sh
	source ./tests/sh/helpers.sh
fi

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

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer -verbose > $outFileName &

	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi

  outputFiles+=("$outFileName")
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
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

./client/client -UIPort=8080 -file=$file_a
./client/client -UIPort=8081 -file=$file_b

sleep 4

./client/client -UIPort=8084 -file=$file_a
./client/client -UIPort=8084 -file=$file_c

sleep 2

./client/client -UIPort=8083 -file=$file_b

sleep 4

pkill -f Peerster

# Tests

echo -e "${NC}# CHECK that everybody has found at least one block${NC}"

expect_contains A "FOUND-BLOCK"
expect_contains B "FOUND-BLOCK"
expect_contains C "FOUND-BLOCK"
expect_contains D "FOUND-BLOCK"
expect_contains E "FOUND-BLOCK"
expect_contains F "FOUND-BLOCK"
expect_contains G "FOUND-BLOCK"
expect_contains H "FOUND-BLOCK"
expect_contains I "FOUND-BLOCK"
expect_contains J "FOUND-BLOCK"

echo -e "${NC}# CHECK that everybody is printing the chain${NC}"

expect_contains A "CHAIN"
expect_contains B "CHAIN"
expect_contains C "CHAIN"
expect_contains D "CHAIN"
expect_contains E "CHAIN"
expect_contains F "CHAIN"
expect_contains G "CHAIN"
expect_contains H "CHAIN"
expect_contains I "CHAIN"
expect_contains J "CHAIN"

echo -e "${NC}# Check that E received the file from A${NC}"

expect_contains A "FORK-LONGER rewind . blocks"
expect_contains B "FORK-LONGER rewind . blocks"
expect_contains C "FORK-LONGER rewind . blocks"
expect_contains D "FORK-LONGER rewind . blocks"
expect_contains E "FORK-LONGER rewind . blocks"
expect_contains F "FORK-LONGER rewind . blocks"
expect_contains G "FORK-LONGER rewind . blocks"
expect_contains H "FORK-LONGER rewind . blocks"
expect_contains I "FORK-LONGER rewind . blocks"
expect_contains J "FORK-LONGER rewind . blocks"

echo -e "${NC}# Check that A, B and E successfully added its file${NC}"

expect_contains A "CANDIDATE transaction $file_a successfily added"
expect_contains B "CANDIDATE transaction $file_b successfily added"
expect_contains E "CANDIDATE transaction $file_c successfily added"

echo -e "${NC}# Check that D and E do not add their duplicate names${NC}"

expect_contains E "IGNORE transaction $file_a already in chain"
expect_contains D "IGNORE transaction $file_b already in chain"

echo -e "${NC}# Check that A,B and E successfully transmitted its file${NC}"

expect_contains A "BROADCAST transaction $file_a"
expect_contains B "BROADCAST transaction $file_b"
expect_contains E "BROADCAST transaction $file_c"

echo -e "${NC}# Check that everybody has the files${NC}"

expect_contains A "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains B "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains C "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains D "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains E "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains F "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains G "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains H "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains I "CHAIN [a-zA-Z0-9 :,.]*$file_a"
expect_contains J "CHAIN [a-zA-Z0-9 :,.]*$file_a"

expect_contains A "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains B "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains C "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains D "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains E "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains F "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains G "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains H "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains I "CHAIN [a-zA-Z0-9 :,.]*$file_b"
expect_contains J "CHAIN [a-zA-Z0-9 :,.]*$file_b"

expect_contains A "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains B "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains C "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains D "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains E "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains F "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains G "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains H "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains I "CHAIN [a-zA-Z0-9 :,.]*$file_c"
expect_contains J "CHAIN [a-zA-Z0-9 :,.]*$file_c"

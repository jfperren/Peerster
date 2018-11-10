#!/usr/bin/env bash

# Same setup as test_routing.sh

go build
cd client
go build
cd ..

# Variables

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="false"

outputFiles=()

UIPort=8080
gossipPort=5000
name="A"
rtimer=1

message_a_c="This is a message from A to C; should be delivered"
message_a_d="This is a message from A to D; should not be delivered"
message_j_a="This is a message from J to A; should be delivered"
message_g_c="This is a message from G to C; should be delivered"
message_j_b="This is a message from J to B; should not be delivered"

# Create Gossipers

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

# Send Private Messages

sleep 2

./client/client -UIPort=8080 -msg="$message_a_c" -dest="C"
./client/client -UIPort=8080 -msg="$message_a_d" -dest="D"
./client/client -UIPort=8089 -msg="$message_j_a" -dest="A"
./client/client -UIPort=8086 -msg="$message_g_c" -dest="C"
./client/client -UIPort=8089 -msg="$message_j_b" -dest="B"

sleep 2

pkill -f Peerster

# Tests

failed="F"

expect_contains() {

  file="logs/${1}.out"
  regex=${2}

  if (grep -q "$regex" ${file}) ; then
    echo -e "${GREEN}- ${file} : <CONTAINS> ${regex}${GREEN}"
  else
    failed="T"
    echo -e "${RED}- ${file} : <SHOULD CONTAIN> ${regex}${RED}"
  fi
}

expect_missing() {

  file="logs/${1}.out"
  regex=${2}

  if !(grep -q "$regex" ${file}) ; then
    echo -e "${GREEN}- ${file} : <DOES NOT CONTAIN> ${regex}${GREEN}"
  else
    failed="T"
    echo -e "${RED}- ${file} : <SHOULD NOT CONTAIN> ${regex}${RED}"
  fi
}

echo -e "${NC}# CHECK that valid recipients got their message${NC}"

expect_contains C "PRIVATE origin A hop-limit "
expect_contains C " contents $message_a_c"
expect_contains A "PRIVATE origin J hop-limit"
expect_contains A " contents $message_j_a"
expect_contains C "PRIVATE origin G hop-limit "
expect_contains C " contents $message_g_c"

echo -e "${NC}# CHECK that other recipients cannot see the messages${NC}"

expect_missing D "$message_a_c"
expect_missing D "$message_j_a"
expect_missing D "$message_g_c"

expect_missing E "$message_a_c"
expect_missing E "$message_j_a"
expect_missing E "$message_g_c"

echo -e "${NC}# CHECK that messages to unknown nodes were not delivered${NC}"

expect_missing D "$message_a_d"
expect_missing B "$message_j_b"

echo -e "${NC}# CHECK that nodes handle unknown routes well${NC}"

expect_contains A "UNKNOWN DESTINATION D"
expect_contains J "UNKNOWN DESTINATION B"

if [[ "$failed" == "T" ]] ; then
	echo -e "${RED}***FAILED***${NC}"
else
	echo -e "${GREEN}***PASSED***${NC}"
fi

#!/usr/bin/env bash

# This script tests the simple case of 5 gossipers sending each other files.
# The setup is a line A - B - C - D - E

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
message_c1_1=Weather_is_clear
message_c2_1=Winter_is_coming
message_c1_2=No_clouds_really
message_c2_2=Let\'s_go_skiing
message_c3=Is_anybody_here?

UIPort=8080
gossipPort=5000
name="A"
rtimer=1

file_a_small="hello.txt"
file_e_big="some_long_message.txt"
file_c_medium=""

downloadDir="_Downloads"
sharedDir="_SharedFiles"

# Start Gossipers & Clean folders

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 5`;
do
	outFileName="logs/$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
  sharedDirName="$sharedDir/$name/"
  downloadDirName="$downloadDir/$name/"

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer> $outFileName &

  rm -rf $downloadDirName && mkdir $downloadDirName
  rm -rf $sharedDirName && mkdir $sharedDirName

	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi

  outputFiles+=("$outFileName")
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

sleep 2

# Create files

touch "$sharedDir/A/$file_a_small"
echo "Hello" > "$sharedDir/A/$file_a_small"

# Clean Alice's downloads

./client/client -UIPort=8080 -file=$file_a_small
./client/client -UIPort=8084 -file=$file_a_small -request=690b8ffe922437a30cb89cab868aee4d1571e625c608f89cec8b6d6380b85af6 -dest="A"

sleep 5

pkill -f Peerster

# Tests

failed="F"

echo -e "${NC}# CHECK that files are scanned correctly${NC}"

expect_contains() {

  file="logs/${1}.out"
  regex=${2}

  if (grep -q "$regex" ${file}) ; then
    echo -e "${GREEN}- ${file} : <CONTAINS> ${regex}${GREEN}"
  else
    failed="T"
    echo -e "${RED}- ${file} : <MISSING> ${regex}${RED}"
  fi
}

expect_contains E "DOWNLOADING metafile of $file_a_small from A"
expect_contains E "DOWNLOADING $file_a_small chunk 0 from A"
expect_contains E "RECONSTRUCTED file $file_a_small"

# if !(grep -q "DOWNLOADING metafile of file $file_a_small from A" "E.out") ; then
#   echo -e "${NC}# CHECK that files are scanned correctly${NC}"
# 	failed="T"
# else;
#   echo -e "${NC}# CHECK that files are scanned correctly${NC}"
# fi
#
# if !(grep -q "DOWNLOADING $file_a_small chunk 0 from A" "E.out") ; then
# 	failed="T"
# fi
#
# if !(grep -q "RECONSTRUCTED file $file_a_small" "E.out") ; then
# 	failed="T"
# fi

if [[ "$failed" == "T" ]] ; then
	echo -e "${RED}***FAILED***${NC}"
else
	echo -e "${GREEN}***PASSED***${NC}"
fi
#
#
# ./client/client -UIPort=12349 -msg=$message_c1_1

# Retrieve file

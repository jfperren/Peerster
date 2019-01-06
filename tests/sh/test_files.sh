#!/usr/bin/env bash

# This script tests the simple case of 5 gossipers sending each other files.
# The setup is a line A - B - C - D - E

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

# Preparation

UIPort=8080
gossipPort=5000
name="A"
rtimer=1

outputFiles=()
file_a_small="hello.txt"
file_b_too_big="too_big.txt"
file_c_medium="message.txt"
file_d_inexistant="inexistant.txt"
file_e_big="poem.txt"
file_f_no_node="no_node.txt"

hash_file_a_small=690b8ffe922437a30cb89cab868aee4d1571e625c608f89cec8b6d6380b85af6
hash_file_b_too_big=0000000000000000000000000000000000000000000000000000000000000000
hash_file_c_medium=f94e8bbcf8808089451e4e3077f79bf8a1b8c0ee08f0fbc738b411278a662a47
hash_file_d_inexistant=0000000000000000000000000000000000000000000000000000000000000000
hash_file_e_big=7995e2cec0eb89ddc7e3642c740d2b4a49de8760d93ccea61e577829b4114369
hash_file_f_no_node=0000000000000000000000000000000000000000000000000000000000000000

downloadDir="_Downloads"
sharedDir="_SharedFiles"

# Start Gossipers & Clean folders

for i in `seq 1 5`;
do
	outFileName="logs/$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
  sharedDirName="$sharedDir/$name/"
  downloadDirName="$downloadDir/$name/"

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer -verbose -separatefs$CRYPTOOPTS > $outFileName &

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

touch "$sharedDir/E/$file_e_big"

for i in `seq 1 20000`;
do
  line="This is an autogenerated poem from a computer, you are reading line $i"
  echo $line >> "$sharedDir/E/$file_e_big"
done

touch "$sharedDir/C/$file_c_medium"

for i in `seq 1 2000`;
do
  line="This is a medium-sized message, line $i"
  echo $line >> "$sharedDir/C/$file_c_medium"
done

touch "$sharedDir/B/$file_b_too_big"

for i in `seq 1 50000`;
do
  line="This is a file that is too big (>2KB) so we can't send. This is line $i"
  echo $line >> "$sharedDir/B/$file_b_too_big"
done

# Upload files & start downloading

./client/client -UIPort=8080 -file=$file_a_small
./client/client -UIPort=8084 -file=$file_e_big
./client/client -UIPort=8081 -file=$file_b_too_big

sleep 2

./client/client -UIPort=8084 -file=$file_a_small -request=$hash_file_a_small -dest="A"
./client/client -UIPort=8080 -file=$file_c_medium -request=$hash_file_c_medium -dest="C"
./client/client -UIPort=8080 -file=$file_d_inexistant -request=$hash_file_d_inexistant -dest="D"
./client/client -UIPort=8080 -file=$file_e_big -request=$hash_file_e_big -dest="E"
./client/client -UIPort=8081 -file=$file_f_no_node -request=$hash_file_f_no_node -dest="F"

sleep 1

./client/client -UIPort=8082 -file=$file_c_medium

sleep 1

./client/client -UIPort=8084 -file=$file_c_medium -request=$hash_file_c_medium -dest="C"

sleep 5

pkill -f Peerster

# Tests

echo -e "${NC}# CHECK that files are scanned correctly${NC}"

# Nothing here

echo -e "${NC}# Check that E received the file from A${NC}"

expect_contains E "DOWNLOADING metafile of $file_a_small from A"
expect_contains E "DOWNLOADING $file_a_small chunk 1 from A"
expect_contains E "RECONSTRUCTED file $file_a_small"

echo -e "${NC}# Check that A received the file from E${NC}"

expect_contains A "DOWNLOADING metafile of $file_e_big from E"
expect_contains A "DOWNLOADING $file_e_big chunk 1 from E"
expect_contains A "DOWNLOADING $file_e_big chunk 50 from E"
expect_contains A "DOWNLOADING $file_e_big chunk 100 from E"
expect_contains A "DOWNLOADING $file_e_big chunk 150 from E"
expect_contains A "DOWNLOADING $file_e_big chunk 180 from E"
expect_contains A "RECONSTRUCTED file $file_e_big"

echo -e "${NC}# Check that A received the file from C${NC}"

expect_contains A "DOWNLOADING metafile of $file_c_medium from C"
expect_contains A "DOWNLOADING $file_c_medium chunk 1 from C"
expect_contains A "DOWNLOADING $file_c_medium chunk 5 from C"
expect_contains A "DOWNLOADING $file_c_medium chunk 10 from C"
expect_contains A "RECONSTRUCTED file $file_c_medium"

echo -e "${NC}# Check that E received the file from C${NC}"

expect_contains E "DOWNLOADING metafile of $file_c_medium from C"
expect_contains E "DOWNLOADING $file_c_medium chunk 1 from C"
expect_contains E "DOWNLOADING $file_c_medium chunk 5 from C"
expect_contains E "DOWNLOADING $file_c_medium chunk 10 from C"
expect_contains E "RECONSTRUCTED file $file_c_medium"

echo -e "${NC}# Check that B did not find F${NC}"

expect_contains B "UNKNOWN DESTINATION F"

echo -e "${NC}# Check that B did not process the file that is too big${NC}"

expect_contains B "WARNING file too_big.txt is too big for Peerster (max. 2Mb)"

echo -e "${NC}# Check that D looked and did not find file asked by A${NC}"

expect_contains D "RECEIVE DATA REQUEST from A to D metahash $hash_file_d_inexistant"
expect_contains D "NOT FOUND hash $hash_file_d_inexistant from A"

if [[ $PACKAGE == false ]]; then
	print_test_results
fi

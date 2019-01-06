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

UIPort=8080
gossipPort=5000
name="A"
rtimer=1

outputFiles=()
file_a="hello.txt"
file_b="poem.txt"
file_c="another_hello.txt"
file_d="my presentation.pdf"
file_e="highway_to_hell.mp3"

hash_a="690b8ffe922437a30cb89cab868aee4d1571e625c608f89cec8b6d6380b85af6"
hash_b="7995e2cec0eb89ddc7e3642c740d2b4a49de8760d93ccea61e577829b4114369"
hash_c="99a5021f52ad2c8981b9caacf1afb48f86d7276d80e0b3636ad92bb08f548fb0"
hash_e="2730f375e4cfbb397bc0f74924ce01f08fc22cd8c5d6df65e1d094a409ebcb97"

chunks_a="1"
chunks_b="1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,
26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,
52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,
78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100,101,102,
103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,
122,123,124,125,126,127,128,129,130,131,132,133,134,135,136,137,138,139,140,
141,142,143,144,145,146,147,148,149,150,151,152,153,154,155,156,157,158,159,160,
161,162,163,164,165,166,167,168,169,170,171,172,173,174,175,176,177,178,179,180"
chunks_c="1,2,3,4,5,6,7,8"
chunks_e="1,2,3,4"

downloadDir="_Downloads"
sharedDir="_SharedFiles"

# Start Gossipers & Clean folders

for i in `seq 1 5`;
do
	outFileName="logs/$name.out"
	peerPort=$((($gossipPort+1)%5+5000))
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

touch "$sharedDir/A/$file_a"
echo "Hello" > "$sharedDir/A/$file_a"

touch "$sharedDir/E/$file_b"

for i in `seq 1 20000`;
do
  line="This is an autogenerated poem from a computer, you are reading line $i"
  echo $line >> "$sharedDir/E/$file_b"
done

touch "$sharedDir/C/$file_c"

for i in `seq 1 2000`;
do
  line="This is an mp3 song, line $i"
  echo $line >> "$sharedDir/C/$file_c"
done

touch "$sharedDir/B/$file_d"

for i in `seq 1 4400`;
do
  line="This is a message that has spaces in the title. This is line $i"
  echo $line >> "$sharedDir/B/$file_d"
done

touch "$sharedDir/E/$file_e"

for i in `seq 1 600`;
do
  line="This is another hello.txt. This is line $i"
  echo $line >> "$sharedDir/E/$file_e"
done

# Upload files & start downloading

./client/client -UIPort=8080 -file="$file_a"
./client/client -UIPort=8084 -file="$file_b"
./client/client -UIPort=8082 -file="$file_c"
./client/client -UIPort=8081 -file="$file_d"
./client/client -UIPort=8084 -file="$file_e"

sleep 3

./client/client -UIPort=8082 -keywords="hell"
./client/client -UIPort=8080 -keywords="inexistant"
./client/client -UIPort=8083 -keywords="txt" -budget 1

sleep 8

pkill -f Peerster

# Tests

echo -e "${NC}# Check that regular searches are started${NC}"

expect_contains A "START search inexistant budget 2 increasing true"
expect_contains C "START search hell budget 2 increasing true"

echo -e "${NC}# Check that D's budget-specific search is started${NC}"

expect_contains D "START search txt budget 1 increasing false"

echo -e "${NC}# Check that C's search got the expected results${NC}"

expect_contains C "FOUND match $file_a at A metafile=$hash_a chunks=$chunks_a"
expect_contains C "FOUND match $file_e at E metafile=$hash_e chunks=$chunks_e"

echo -e "${NC}# Check that D's search got the expected results${NC}"

expect_contains D "FOUND match $file_c at C metafile=$hash_c chunks=$chunks_c"
expect_contains D "FOUND match $file_b at E metafile=$hash_b chunks=$chunks_b"

echo -e "${NC}# Check that C and D's search completed${NC}"

expect_contains C "SEARCH FINISHED"
expect_contains D "SEARCH FINISHED"

echo -e "${NC}# Check that C's search does not match itself${NC}"

expect_missing A "FOUND match $file_c at C"

echo -e "${NC}# Check that D's search did not match A's file (too far)${NC}"

expect_contains C "FOUND match $file_a at A metafile=$hash_a chunks=$chunks_a"
expect_contains C "FOUND match $file_e at E metafile=$hash_e chunks=$chunks_e"

echo -e "${NC}# Check that A's inexistant search does not match anything${NC}"

expect_missing A "FOUND match"

echo -e "${NC}# Check that A's search increases as expected${NC}"

expect_contains A "START search inexistant budget 4 increasing true"
expect_contains A "START search inexistant budget 8 increasing true"
expect_contains A "START search inexistant budget 16 increasing true"
expect_contains A "START search inexistant budget 32 increasing true"

echo -e "${NC}# Check that A's search stops after 32${NC}"

expect_contains A "TIMEOUT search inexistant"
expect_missing A "START search inexistant budget 64 increasing true"

# CHECK THAT NODES WITH HALF OF THINGS CAN STILL REPLY

if [[ $PACKAGE == false ]]; then
	print_test_results
fi

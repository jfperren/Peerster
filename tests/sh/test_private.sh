#!/usr/bin/env bash

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
message_a_d="This is a message from A to D; should not be delivered"
message_j_a="This is a message from J to A; should be delivered"
message_g_c="This is a message from G to C; should be delivered"
message_j_b="This is a message from J to B; should not be delivered"

# Create Gossipers

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

	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=$rtimer -verbose$CRYPTOOPTS 2> $outFileName &

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

if [[ $nb_nodes > 2 ]]
then
    ./client/client -UIPort=8080 -msg="$message_a_c" -dest="C"
fi
if [[ $nb_nodes > 3 ]]
then
    ./client/client -UIPort=8080 -msg="$message_a_d" -dest="D"
fi
if [[ $nb_nodes > 9 ]]
then
    ./client/client -UIPort=8089 -msg="$message_j_a" -dest="A"
fi
if [[ $nb_nodes > 6 ]]
then
    ./client/client -UIPort=8086 -msg="$message_g_c" -dest="C"
fi
if [[ $nb_nodes > 9 ]]
then
    ./client/client -UIPort=8089 -msg="$message_j_b" -dest="B"
fi

sleep 2

pkill -f Peerster

# Tests

echo -e "${NC}# CHECK that valid recipients got their message${NC}"

if [[ $nb_nodes > 2 ]]
then
    expect_contains C "PRIVATE origin A hop-limit "
    expect_contains C " contents $message_a_c"
fi
if [[ $nb_nodes > 9 ]]
then
    expect_contains A "PRIVATE origin J hop-limit"
    expect_contains A " contents $message_j_a"
fi
if [[ $nb_nodes > 6 ]]
then
    expect_contains C "PRIVATE origin G hop-limit "
    expect_contains C " contents $message_g_c"
fi

echo -e "${NC}# CHECK that other recipients cannot see the messages${NC}"

if [[ $nb_nodes > 3 ]]
then
    expect_missing D "$message_a_c"
    if [[ $nb_nodes > 9 ]]
    then
        expect_missing D "$message_j_a"
    fi
    if [[ $nb_nodes > 6 ]]
    then
        expect_missing D "$message_g_c"
    fi
fi

if [[ $nb_nodes > 4 ]]
then
    expect_missing E "$message_a_c"
    if [[ $nb_nodes > 9 ]]
    then
        expect_missing E "$message_j_a"
    fi
    if [[ $nb_nodes > 6 ]]
    then
        expect_missing E "$message_g_c"
    fi
fi

echo -e "${NC}# CHECK that messages to unknown nodes were not delivered${NC}"

if [[ $nb_nodes > 3 ]]
then
    expect_missing D "$message_a_d"
fi
if [[ $nb_nodes > 9 ]]
then
    expect_missing B "$message_j_b"
fi

echo -e "${NC}# CHECK that nodes handle unknown routes well${NC}"

if [[ $nb_nodes > 3 ]]
then
expect_contains A "UNKNOWN DESTINATION D"
fi
if [[ $nb_nodes > 9 ]]
then
expect_contains J "UNKNOWN DESTINATION B"
fi

if [[ $PACKAGE == false ]]; then
	print_test_results
fi

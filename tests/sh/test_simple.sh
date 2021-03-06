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

# Preparation

outputFiles=()
message=Weather_is_clear
message2=Winter_is_coming

UIPort=12345
gossipPort=5000
name='A'

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipPort=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 $nb_nodes`;
do
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%$nb_nodes+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -simple -peers=$peer$CRYPTOOPTS 2> $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

./client/client -UIPort=12349 -msg=$message
./client/client -UIPort=12346 -msg=$message2
sleep 3
pkill -f Peerster

# Testing

if [[ $nb_nodes > 4 ]] && !(grep -q "CLIENT MESSAGE $message" "E.out") ; then
	failed="T"
fi

if !(grep -q "CLIENT MESSAGE $message2" "B.out") ; then
  failed="T"
fi

if [[ "$failed" == "T" ]] ; then
	echo -e "${RED}FAILED${NC}"
fi

gossipPort=5000
for i in `seq 0 $nb_nodes`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$((4999 + $nb_nodes))
	fi
	nextPort=$((($gossipPort+1)%$nb_nodes+5000))
	msgLine="SIMPLE MESSAGE origin E from 127.0.0.1:$relayPort contents $message"
	msgLine2="SIMPLE MESSAGE origin B from 127.0.0.1:$relayPort contents $message2"
	peersLine="127.0.0.1:$nextPort,127.0.0.1:$relayPort"
	if [[ "$DEBUG" == "true" ]] ; then
		echo "check 1 $msgLine"
		echo "check 2 $msgLine2"
		echo "check 3 $peersLine"
	fi
	gossipPort=$(($gossipPort+1))
	if [[ $nb_nodes > 4 ]] && !(grep -q "$msgLine" "${outputFiles[$i]}") ; then
   		failed="T"
	fi
	if !(grep -q "$peersLine" "${outputFiles[$i]}") ; then
        failed="T"
    fi
	if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
        failed="T"
    fi
done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
	echo "***PASSED***"
fi

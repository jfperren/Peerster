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
message_c1_1=Weather_is_clear
message_c2_1=Winter_is_coming
message_c1_2=No_clouds_really
message_c2_2=Let\'s_go_skiing
message_c3=Is_anybody_here?

UIPort=12345
gossipPort=5000
name='A'

for i in `seq 1 $nb_nodes`;
do
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%$nb_nodes+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer$CRYPTOOPTS 2> $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

if [[ $nb_nodes > 4 ]]
then
./client/client -UIPort=12349 -msg=$message_c1_1
fi
./client/client -UIPort=12346 -msg=$message_c2_1
sleep 2

if [[ $nb_nodes > 4 ]]
then
./client/client -UIPort=12349 -msg=$message_c1_2
sleep 1
fi
./client/client -UIPort=12346 -msg=$message_c2_2
if [[ $nb_nodes > 6 ]]
then
./client/client -UIPort=12351 -msg=$message_c3
fi

sleep 5
pkill -f Peerster

# Testing

echo -e "${RED}###CHECK that client messages arrived${NC}"

if [[ $nb_nodes > 4 ]]
then
    if !(grep -q "CLIENT MESSAGE $message_c1_1" "E.out") ; then
        failed="T"
    fi

    if !(grep -q "CLIENT MESSAGE $message_c1_2" "E.out") ; then
        failed="T"
    fi
fi

if !(grep -q "CLIENT MESSAGE $message_c2_1" "B.out") ; then
    failed="T"
fi

if !(grep -q "CLIENT MESSAGE $message_c2_2" "B.out") ; then
    failed="T"
fi

if [[ $nb_nodes > 6 ]]
then
    if !(grep -q "CLIENT MESSAGE $message_c3" "G.out") ; then
        failed="T"
    fi
fi

if [[ "$failed" == "T" ]] ; then
	echo -e "${RED}***FAILED***${NC}"
else
	echo -e "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo -e "${RED}###CHECK rumor messages ${NC}"

gossipPort=5000
for i in `seq 0 $(($nb_nodes - 1))`;
do
	relayPort=$(($gossipPort-1))
	if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$((4999 + $nb_nodes))
	fi
	nextPort=$((($gossipPort+1)%$nb_nodes+5000))
	msgLine1="RUMOR origin E from 127.0.0.1:[0-9]{4} ID 1 contents $message_c1_1"
	msgLine2="RUMOR origin E from 127.0.0.1:[0-9]{4} ID 2 contents $message_c1_2"
	msgLine3="RUMOR origin B from 127.0.0.1:[0-9]{4} ID 1 contents $message_c2_1"
	msgLine4="RUMOR origin B from 127.0.0.1:[0-9]{4} ID 2 contents $message_c2_2"
	msgLine5="RUMOR origin G from 127.0.0.1:[0-9]{4} ID 1 contents $message_c3"

	if [[ "$gossipPort" != 5004 ]] && [[ $nb_nodes > 4 ]]; then
		if !(grep -Eq "$msgLine1" "${outputFiles[$i]}") ; then
        	failed="T"
    	fi
		if !(grep -Eq "$msgLine2" "${outputFiles[$i]}") ; then
        	failed="T"
    	fi
	fi

	if [[ "$gossipPort" != 5001 ]] ; then
		if !(grep -Eq "$msgLine3" "${outputFiles[$i]}") ; then
        	failed="T"
    	fi
		if !(grep -Eq "$msgLine4" "${outputFiles[$i]}") ; then
        	failed="T"
    	fi
	fi

	if [[ "$gossipPort" != 5006 ]] && [[ $nb_nodes > 6 ]] ; then
		if !(grep -Eq "$msgLine5" "${outputFiles[$i]}") ; then
        	failed="T"
    	fi
	fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
    echo -e "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo -e "${RED}###CHECK mongering${NC}"
gossipPort=5000
for i in `seq 0 $(($nb_nodes - 1))`;
do
    relayPort=$(($gossipPort-1))
    if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$(($nb_nodes + 4999))
    fi
    nextPort=$((($gossipPort+1)%$nb_nodes+5000))

    msgLine1="MONGERING with 127.0.0.1:$relayPort"
    msgLine2="MONGERING with 127.0.0.1:$nextPort"

    if !(grep -q "$msgLine1" "${outputFiles[$i]}") && !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
    echo -e "${GREEN}***PASSED***${NC}"
fi


failed="F"
echo -e "${RED}###CHECK status messages ${NC}"
gossipPort=5000
for i in `seq 0 $(($nb_nodes - 1))`;
do
    relayPort=$(($gossipPort-1))
    if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$(($nb_nodes + 4999))
    fi
    nextPort=$((($gossipPort+1)%$nb_nodes+5000))

	msgLine1="STATUS from 127.0.0.1:$relayPort"
	msgLine2="STATUS from 127.0.0.1:$nextPort"
	msgLine3="peer E nextID 3"
	msgLine4="peer B nextID 3"
	msgLine5="peer G nextID 2"

	if !(grep -q "$msgLine1" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    if [[ $nb_nodes > 4 ]] && !(grep -q "$msgLine3" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    if !(grep -q "$msgLine4" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    if [[ $nb_nodes > 6 ]] && !(grep -q "$msgLine5" "${outputFiles[$i]}") ; then
        failed="T"
    fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
    echo -e "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo -e "${RED}###CHECK flipped coin${NC}"
gossipPort=5000
for i in `seq 0 $(($nb_nodes - 1))`;
do
    relayPort=$(($gossipPort-1))
    if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$(($nb_nodes + 4999))
    fi
    nextPort=$((($gossipPort+1)%$nb_nodes+5000))

    msgLine1="FLIPPED COIN sending rumor to 127.0.0.1:$relayPort"
    msgLine2="FLIPPED COIN sending rumor to 127.0.0.1:$nextPort"

    if !(grep -q "$msgLine1" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
        failed="T"
    fi
	gossipPort=$(($gossipPort+1))

done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
    echo -e "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo -e "${RED}###CHECK in sync${NC}"
gossipPort=5000
for i in `seq 0 $(($nb_nodes - 1))`;
do
    relayPort=$(($gossipPort-1))
    if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$(($nb_nodes + 4999))
    fi
    nextPort=$((($gossipPort+1)%$nb_nodes+5000))

    msgLine1="IN SYNC WITH 127.0.0.1:$relayPort"
    msgLine2="IN SYNC WITH 127.0.0.1:$nextPort"

    if !(grep -q "$msgLine1" "${outputFiles[$i]}") ; then
        failed="T"
    fi
    if !(grep -q "$msgLine2" "${outputFiles[$i]}") ; then
        failed="T"
    fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
    echo -e "${GREEN}***PASSED***${NC}"
fi

failed="F"
echo -e "${RED}###CHECK correct peers${NC}"
gossipPort=5000
for i in `seq 0 $(($nb_nodes - 1))`;
do
    relayPort=$(($gossipPort-1))
    if [[ "$relayPort" == 4999 ]] ; then
        relayPort=$(($nb_nodes + 4999))
    fi
    nextPort=$((($gossipPort+1)%$nb_nodes+5000))

	peersLine1="127.0.0.1:$relayPort,127.0.0.1:$nextPort"
	peersLine2="127.0.0.1:$nextPort,127.0.0.1:$relayPort"

    if !(grep -q "$peersLine1" "${outputFiles[$i]}") && !(grep -q "$peersLine2" "${outputFiles[$i]}") ; then
        failed="T"
    fi
	gossipPort=$(($gossipPort+1))
done

if [[ "$failed" == "T" ]] ; then
    echo -e "${RED}***FAILED***${NC}"
else
    echo -e "${GREEN}***PASSED***${NC}"
fi

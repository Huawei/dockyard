#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

#set default value of args
IP="0.0.0.0"
PORT="9632"
MASTERIP=""
MASTERPORT="8099"
CHUNKNUM="2"
GROUPID="1"
DATADIR="./DATA"
ERRLOGDIR="."

while getopts ":i:p:m:n:c:g:d:e:" args 
do
	case $args in
		i)
            IP=$OPTARG
			echo "IP=$IP"
			;;
		p)
            PORT=$OPTARG
			echo "PORT=$PORT"
			;;
		m)
		    MASTERIP=$OPTARG
			echo "MASTERIP=$MASTERIP"
			;;
		n)
		    MASTERPORT=$OPTARG
			echo "MASTERPORT=$MASTERPORT"
			;;
		c)
			CHUNKNUM=$OPTARG
			echo "CHUNKNUM=$CHUNKNUM"
			;;
		g)
			GROUPID=$OPTARG
			echo "GROUPID=$GROUPID"
			;;
		d)
			DATADIR=$OPTARG
			echo "DATADIR=$DATADIR"
			;;
		e)
			ERRLOGDIR=$OPTARG
			echo "ERRLOGDIR=$ERRLOGDIR"
			;;
		"?")  
      		echo "Invalid option: -$OPTARG"   
      		;;  
      	":")
        	echo "No argument value for option $OPTARG"
        	;;
	esac
done


if [ -d "$DATADIR" ]; then
	echo "DATA folder '$DATADIR' exist"
else
	echo "DATADIR NOT EXIST, CREATE IT"
	mkdir -p "$DATADIR"
fi

if [ -d "$ERRLOGDIR" ]; then
	echo "Error log folder '$ERRLOGDIR' exist"
else
	echo "ERRLOGDIR NOT EXIST, CREATE IT"
	mkdir -p "$ERRLOGDIR"
fi

ERRLOGPATH=${ERRLOGDIR}/error.log

./oss/chunkserver/spy_server --ip $IP --port $PORT --master_ip $MASTERIP --master_port $MASTERPORT --chunks $CHUNKNUM --group_id $GROUPID --data_dir $DATADIR --error_log $ERRLOGPATH
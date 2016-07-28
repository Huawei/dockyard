#!/bin/sh
set -e

echo "start test, killing the exist 'dus' server"
pidof dus | xargs kill -9

TMPDIR=$(mktemp -d)
TMPSTORAGE=${TMPDIR}/storage
TMPKM=${TMPDIR}/km
echo "creating tmp storage dir: " $TMPSTORAGE
mkdir -p $TMPSTORAGE
echo "creating tmp keymanager dir: " $TMPKM
mkdir -p $TMPKM


echo "start to compile server"
cd server
make

echo "start the updater server"
./dus web --storage "local:/""$TMPSTORAGE" --keymanager "local:/""$TMPKM"  &

echo "start to compile client"
cd ../client
make
cd ..

echo "set enviornment and start to run tests"
export DUS_TEST_SERVER="appV1://localhost:1234"
go test -v $(go list ./... | grep -v /vendor/)

echo "killing the testing 'dus' server"
killall dus

echo "clean all the generated data"
rm -fr $TMPDIR

echo "end of test"

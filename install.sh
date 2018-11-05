#!/bin/bash
VERSION=$( go version )
NUMBER=$( echo -n "$VERSION" | grep -o "[0-9]\.[0-9]*" )
REQUIREDMAJORVERSION=1
REQUIREDMINORVERSION=10
MAJORVERSION=$( echo -n "$NUMBER" | cut -d "." -f 1 )
MINORVERSION=$( echo -n "$NUMBER" | cut -d "." -f 2 )
if [ "$( echo "${MAJORVERSION} < ${REQUIREDMAJORVERSION}" | bc )" == 1 ] ; then
	echo "Go version $MAJORVERSION.$MINORVERSION is not supported, please install go version $REQUIREDMAJORVERSION.$REQUIREDMINORVERSION or newer"
	exit 1
fi
if [ "$( echo "${MINORVERSION} < ${REQUIREDMINORVERSION}" | bc )" == 1 ] ; then
	echo "Go version $MAJORVERSION.$MINORVERSION is not supported, please install go version $REQUIREDMAJORVERSION.$REQUIREDMINORVERSION or newer"
	exit 1
fi

./build.sh

go get -u github.com/aws/aws-sdk-go/... 
go get github.com/google/go-cmp/cmp 
go get github.com/sevlyar/go-daemon 



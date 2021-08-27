#!/bin/bash

function build(){
    set -e
    for microDir in micro-api micro-cluster micro-manager micro-metadb
    do
        echo "build $microDir ..."
        cd $microDir
        go build .
        if [ "$microDir" == "micro-cluster" ]
        then
            echo "    build tiupmgr ..."
            cd tiupmgr
            go build .
            cd ..
        fi
        cd ..
    done
    set +e
}

# $1: install dir
function install(){
    set -e
    installDir=$1
    srcDir=`pwd`
    mkdir -p "$installDir"
    cd "$installDir"
    for microDir in micro-api micro-cluster micro-manager micro-metadb
    do
        mkdir -p "$microDir"
        binName="$microDir"
        cp -fr "$srcDir"/"$microDir"/"$binName" "$microDir"
        cp -fr "$srcDir"/"$microDir"/"cfg.toml" "$microDir"
        if [ "$microDir" == "micro-cluster" ]
        then
            cd "$microDir"
            mkdir -p "tiupmgr"
            cp -fr "$srcDir"/"$microDir""/tiupmgr/tiupmgr" tiupmgr
            cd ..
        fi
    done
    set +e
}

function clean(){
    for microDir in micro-api micro-cluster micro-manager micro-metadb
    do
        echo "clean $microDir ..."
        cd $microDir
        binName=$microDir
        rm -f $binName
        if [ "$microDir" == "micro-cluster" ]
        then
            echo "clean tiupmgr ..."
            cd tiupmgr
            rm -f tiupmgr
            cd ..
        fi
        cd ..
    done
}

function test(){
    exit 1
}

function unitTest(){
    set -e
    find . -iname "*_test.go" -exec dirname {} \; | sort -u | sed -e "s/^\./github.com\/pingcap\/ticp/" | while read pkg
    do
        echo "unit test:$pkg"
        go test -race -cover "$pkg"
    done
    set +e
}

option=$1
arg2=$2
case $option in
  build)
    echo build
    build
    ;;
  install)
    echo install
    installDir="$arg2"
    if [ -z "$installDir" ]
    then
        installDir=bin
    fi
    install "$installDir"
    ;;
  clean)
    echo clean
    clean
    ;;
  test)
    echo test
    unitTest
    ;;
  *)
    echo -n "unknown option $option"
    exit 1
    ;;
esac


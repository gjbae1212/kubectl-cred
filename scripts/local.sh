#!/bin/bash

set -e -o pipefail

trap '[ "$?" -eq 0 ] || echo "Error Line:<$LINENO> Error Function:<${FUNCNAME}>"' EXIT

cd `dirname $0` && cd ..
CURRENT=`pwd`

function test
{
   set_env
   go test -v $(go list ./... | grep -v vendor) --count 1 -covermode=atomic -timeout 120s
}

function set_env
{
  if [[ -f $CURRENT/scripts/local_env.sh ]];
  then
    source $CURRENT/scripts/local_env.sh
  fi
}

CMD=$1
shift
$CMD $*


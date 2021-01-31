#!/bin/bash

set -e -o pipefail

trap '[ "$?" -eq 0 ] || echo "Error Line:<$LINENO> Error Function:<${FUNCNAME}>"' EXIT

cd `dirname $0` && cd ..
CURRENT=`pwd`

function make_krew
{
  # write result to cred.yaml.
  docker run -v $CURRENT/.krew.yaml:/tmp/template-file.yaml rajatjindal/krew-release-bot:v0.0.38 \
  krew-release-bot template --tag v0.1.4 --template-file /tmp/template-file.yaml
}

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


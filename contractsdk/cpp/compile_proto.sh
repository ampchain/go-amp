#!/bin/bash

set -e

cd `dirname $0`
protoc --cpp_out=AmpChain -I ../pb ../pb/contract.proto

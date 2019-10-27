#!/bin/bash

# install docker in precondition

docker run -u $UID --rm -v $(pwd):/src hub.ampio/AmpChain/emcc emmake make

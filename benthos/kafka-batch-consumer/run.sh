#!/bin/bash

# please do `chmod +x run.sh` on this file before run
export $(grep -v '^#' .env | xargs -0)
echo "$ES_URL"
benthos -c config.yaml -r "resources/*.yaml" streams streams/*.yaml


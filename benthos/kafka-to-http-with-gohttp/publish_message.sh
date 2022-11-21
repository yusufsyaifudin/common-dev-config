#!/bin/bash

# please do `chmod +x publish_message.sh` and `chmod 755 publish_message.sh` on this file before run
export $(grep -v '^#' .env | xargs -0)
echo "Bash version ${BASH_VERSION}..."

# produce unique message:
# ensure that same id goes to same partition
for id in {1..100}
do
    interval=$((id%3))

    if [[ interval -eq 0 ]]
    then
       msg="{\"http_status\":200,\"sleep_for\":\"2s\"}"
    else
       msg="{\"http_status\":500,\"sleep_for\":\"5s\"}"
    fi

    echo "$msg"
    echo "$msg" | kcat -P -b "${KAFKA_BROKER_1}" -t "${KAFKA_TOPIC}" -p 0
done
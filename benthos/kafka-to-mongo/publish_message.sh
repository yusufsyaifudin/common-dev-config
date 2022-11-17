#!/bin/bash

# please do `chmod +x publish_message.sh` and `chmod 755 publish_message.sh` on this file before run
export $(grep -v '^#' .env | xargs -0)
echo "Bash version ${BASH_VERSION}..."

# produce unique message:
# ensure that same `service-{num}` goes to same partition
for id in {1..3}
do
  for status in "sent" "delivered" "read"
  do
    partition=$((id%3))
    msg="{\"header\":{\"target_collection\":\"service-${partition}\"},\"body\":{\"key\":\"data-${id}\",\"another_key\":{\"foo\":\"bar\"}}}"
    echo "$msg"
    echo "$msg" | kcat -P -b "${KAFKA_BROKER_1}" -t "${KAFKA_TOPIC}" -p "${partition}"
  done

done
#!/bin/sh
set -e

BOOTSTRAP="${KAFKA_BOOTSTRAP:-kafka:9092}"

topics="
order.created
payment.processed
payment.failed
order.paid
order.ready
courier.assigned
order.delivered
order.status.changed
"

for topic in $topics; do
  /opt/kafka/bin/kafka-topics.sh --bootstrap-server "$BOOTSTRAP" \
    --create --if-not-exists --topic "$topic" \
    --partitions 1 --replication-factor 1
  echo "topic ready: $topic"
done

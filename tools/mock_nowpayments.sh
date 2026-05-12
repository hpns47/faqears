#!/usr/bin/env bash
set -euo pipefail

PAYMENT_ID=""
STATUS=""
IPN_SECRET=""
GATEWAY_URL="http://localhost:8080/api/v1/payments/nowpayments/webhook"

for arg in "$@"; do
  case "$arg" in
    --payment-id=*) PAYMENT_ID="${arg#*=}" ;;
    --status=*)     STATUS="${arg#*=}" ;;
    --ipn-secret=*) IPN_SECRET="${arg#*=}" ;;
    --url=*)        GATEWAY_URL="${arg#*=}" ;;
  esac
done

if [[ -z "$PAYMENT_ID" || -z "$STATUS" || -z "$IPN_SECRET" ]]; then
  echo "Usage: $0 --payment-id=<uuid> --status=<status> --ipn-secret=<secret> [--url=<webhook-url>]"
  exit 1
fi

NOW="$(date -u +%s)"

BODY=$(printf '{"order_id":"%s","payment_id":"%s","payment_status":"%s","created_nomally":true,"occurred_at":%s}' \
  "$PAYMENT_ID" "$PAYMENT_ID" "$STATUS" "$NOW")

SORTED_BODY=$(echo "$BODY" | python3 -c "
import json, sys
data = json.load(sys.stdin)
keys = sorted(data.keys())
parts = []
for k in keys:
    parts.append(json.dumps(k) + ':' + json.dumps(data[k], separators=(',',':')))
print('{' + ','.join(parts) + '}', end='')
")

SIG=$(printf '%s' "$SORTED_BODY" | openssl dgst -sha512 -hmac "$IPN_SECRET" | awk '{print $2}')

curl -sS -X POST "$GATEWAY_URL" \
  -H "Content-Type: application/json" \
  -H "x-nowpayments-sig: $SIG" \
  -d "$BODY"
echo

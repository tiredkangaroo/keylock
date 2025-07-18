
#!/bin/sh

echo "Waiting for Vault..."
until vault status | grep -q 'Initialized.*true'; do
  sleep 1
done

echo "Enabling KV v2 at keylock path..."
vault secrets enable -path=keylock kv-v2 || echo "KV engine at keylock already enabled"

echo "Vault is ready. Writing secrets..."

: "${ENC_KEY:?Need to set ENC_KEY}"

if ! echo "$ENC_KEY" | grep -Eq '^[a-fA-F0-9]{64}$'; then
  echo "ENC_KEY must be exactly 64 hex characters"
  exit 1
fi

if [ -n "$REDIS_USER" ] || [ -n "$REDIS_PASS" ]; then
  vault kv put keylock/redis \
    ${REDIS_USER:+username="$REDIS_USER"} \
    ${REDIS_PASS:+password="$REDIS_PASS"}
fi

if [ -n "$PSQL_USER" ] || [ -n "$PSQL_PASS" ]; then
  vault kv put keylock/psql \
    ${PSQL_USER:+username="$PSQL_USER"} \
    ${PSQL_PASS:+password="$PSQL_PASS"}
fi

vault kv put keylock/encryption enc_key="$ENC_KEY"

echo "Secrets written to Vault."

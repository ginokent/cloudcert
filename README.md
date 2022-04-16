# cloudacme

## Check and update Certificate

```bash
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{
  "vaultProvider": "gcloud",
  "privateKeyVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-key",
  "certificateVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-crt",
  "dnsProvider": "gcloud",
  "termsOfServiceAgreed": true,
  "dnsProviderID": "'${GOOGLE_CLOUD_PROJECT:?}'",
  "email": "'${EMAIL:?}'",
  "thresholdOfDaysToExpire": 30,
  "domains": [
    "'${DOMAIN:?}'",
    "*.'${DOMAIN:?}'"
  ],
  "staging": true
}'

# one-line
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{"vaultProvider":"gcloud","privateKeyVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-key","certificateVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-crt","dnsProvider":"gcloud","termsOfServiceAgreed":true,"dnsProviderID":"'${GOOGLE_CLOUD_PROJECT:?}'","email":"'${EMAIL:?}'","thresholdOfDaysToExpire":30,"domains":["'${DOMAIN:?}'","*.'${DOMAIN:?}'"],"staging":true}'
```

## Force to update Private Key and Certificate

```bash
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{
  "vaultProvider": "gcloud",
  "privateKeyVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-key",
  "certificateVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-crt",
  "renewPrivateKey":true,
  "dnsProvider": "gcloud",
  "termsOfServiceAgreed": true,
  "dnsProviderID": "'${GOOGLE_CLOUD_PROJECT:?}'",
  "email": "'${EMAIL:?}'",
  "thresholdOfDaysToExpire": 30,
  "domains": [
    "'${DOMAIN:?}'",
    "*.'${DOMAIN:?}'"
  ],
  "staging": true
}'

# one-line
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{"vaultProvider":"gcloud","privateKeyVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-key","certificateVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-crt","dnsProvider":"gcloud","termsOfServiceAgreed":true,"dnsProviderID":"'${GOOGLE_CLOUD_PROJECT:?}'","email":"'${EMAIL:?}'","thresholdOfDaysToExpire":30,"domains":["'${DOMAIN:?}'","*.'${DOMAIN:?}'"],"staging":true,"renewPrivateKey":true}'
```

# cloudacme

## Check and update Certificate

```bash
export GOOGLE_CLOUD_PROJECT=**** EMAIL=**** DOMAIN=****
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{
  "vaultProvider": "gcloud",
  "acmeAccountKeyVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-accountsecret",
  "privateKeyVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-privatekey",
  "certificateVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-certificate",
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
export GOOGLE_CLOUD_PROJECT=**** EMAIL=**** DOMAIN=****
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{"vaultProvider":"gcloud","acmeAccountKeyVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-accountsecret","privateKeyVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-privatekey","certificateVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-certificate","dnsProvider":"gcloud","termsOfServiceAgreed":true,"dnsProviderID":"'${GOOGLE_CLOUD_PROJECT:?}'","email":"'${EMAIL:?}'","thresholdOfDaysToExpire":30,"domains":["'${DOMAIN:?}'","*.'${DOMAIN:?}'"],"staging":true}'
```

## Force to update Private Key and Certificate

```bash
export GOOGLE_CLOUD_PROJECT=**** EMAIL=**** DOMAIN=****
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{
  "vaultProvider": "gcloud",
  "acmeAccountKeyVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-accountsecret",
  "privateKeyVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-privatekey",
  "certificateVaultResource": "projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-certificate",
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
export GOOGLE_CLOUD_PROJECT=**** EMAIL=**** DOMAIN=****
curl -i -X POST http://localhost:8080/rpc/v1/certificates/issue -w '\n' -d '{"vaultProvider":"gcloud","acmeAccountKeyVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-accountsecret","privateKeyVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-privatekey","certificateVaultResource":"projects/'${GOOGLE_CLOUD_PROJECT:?}'/secrets/test-cloudacme-certificate","dnsProvider":"gcloud","termsOfServiceAgreed":true,"dnsProviderID":"'${GOOGLE_CLOUD_PROJECT:?}'","email":"'${EMAIL:?}'","thresholdOfDaysToExpire":30,"domains":["'${DOMAIN:?}'","*.'${DOMAIN:?}'"],"staging":true,"renewPrivateKey":true}'
```

package main

import (
	"log"

	cloudcertv1 "github.com/ginokent/cloudcert/generated/go/proto/v1/cloudcert"
)

func main() {
	log.Printf("%#v\n", cloudcertv1.DNSProvider_DNS_PROVIDER_GCLOUD.Descriptor())
}

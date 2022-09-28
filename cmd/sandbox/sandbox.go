package main

import (
	"log"

	cloudcertv1 "github.com/ginokent/cloudcert/generated/go/proto/v1/cloudcert"
	"github.com/kunitsuinc/util.go/protoext"
)

func main() {
	log.Printf("%s %s\n", cloudcertv1.DNSProvider_DNS_PROVIDER_GCLOUD, protoext.EnumValueOptionsString(cloudcertv1.E_EnumStringer, cloudcertv1.DNSProvider_DNS_PROVIDER_GCLOUD))
}

package main

import (
	"log"

	cloudcertv1 "github.com/ginokent/cloudcert/generated/go/proto/v1/cloudcert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

func main() {
	log.Printf(
		"%#v\n",
		proto.GetExtension(
			cloudcertv1.DNSProvider_DNS_PROVIDER_GCLOUD.Descriptor().Values().Get(int(cloudcertv1.DNSProvider_DNS_PROVIDER_GCLOUD)).Options(),
			cloudcertv1.E_EnumStringer,
		),
	)

	log.Printf(
		"%s\n",
		ProtoEnumString(cloudcertv1.DNSProvider_DNS_PROVIDER_GCLOUD, cloudcertv1.E_EnumStringer),
	)
}

type ProtoEnumStringer interface {
	~int32
	Descriptor() protoreflect.EnumDescriptor
	String() string
}

func ProtoEnumString[T ProtoEnumStringer](d T, info *protoimpl.ExtensionInfo) string {
	ext, ok := proto.GetExtension(d.Descriptor().Values().Get(int(d)).Options(), info).(string)

	if !ok {
		return d.String()
	}

	return ext
}

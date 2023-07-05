package marshaler

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

func JsonMarshaler() runtime.Marshaler {
	return &runtime.HTTPBodyMarshaler{
		Marshaler: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseEnumNumbers:  true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	}
}

func ProtoMarshaler() runtime.Marshaler {
	return &runtime.HTTPBodyMarshaler{
		Marshaler: &runtime.ProtoMarshaller{},
	}
}

func GetMarshaler(accept string) runtime.Marshaler {
	switch accept {
	case "application/octet-stream":
		return ProtoMarshaler()
	default:
		return JsonMarshaler()
	}
}

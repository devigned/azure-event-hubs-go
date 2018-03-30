package eventhub

import (
	"context"
	"encoding/binary"
	"net"

	"github.com/opentracing/opentracing-go"
	tag "github.com/opentracing/opentracing-go/ext"
)

func (h *Hub) startSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	tag.Component.Set(span, "eventhubs-go")
	tag.SpanKindRPCClient.Set(span)
	span, err := addIPAddressesToSpan(span)
	return span, ctx, err
}

func (s *sender) startProducerSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	tag.SpanKindRPCClient.Set(span)
	tag.SpanKindProducer.Set(span)
	tag.MessageBusDestination.Set(span, s.getAddress())
	return span, ctx
}

func (r *receiver) startConsumerSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	tag.SpanKindRPCClient.Set(span)
	tag.SpanKindConsumer.Set(span)
	tag.MessageBusDestination.Set(span, r.getAddress())
	return span, ctx
}

func addIPAddressesToSpan(span opentracing.Span) (opentracing.Span, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			switch {
			case ip.To4() != nil:
				tag.PeerHostIPv4.Set(span, ip2int(ip))
			case ip.To16() != nil:
				tag.PeerHostIPv6.Set(span, ip.String())
			}
		}
	}
	return span, nil
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

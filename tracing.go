package eventhub

import (
	"context"
	"encoding/binary"
	"net"
	"os"

	"github.com/opentracing/opentracing-go"
	tag "github.com/opentracing/opentracing-go/ext"
)

func (h *Hub) startSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	tryApplyCommonInfo(span)
	tag.SpanKindRPCClient.Set(span)
	tryApplyNetworkInfo(span)
	return span, ctx
}

func (s *sender) startProducerSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	tryApplyCommonInfo(span)
	tag.SpanKindRPCClient.Set(span)
	tag.SpanKindProducer.Set(span)
	tag.MessageBusDestination.Set(span, s.getFullIdentifier())
	tryApplyNetworkInfo(span)
	return span, ctx
}

func (r *receiver) startConsumerSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, operationName)
	tryApplyCommonInfo(span)
	tag.SpanKindRPCClient.Set(span)
	tag.SpanKindConsumer.Set(span)
	tag.MessageBusDestination.Set(span, r.getFullIdentifier())
	return span, ctx
}

func (r *receiver) startConsumerSpanFromContextFollowing(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span := opentracing.StartSpan(operationName, opentracing.FollowsFrom(opentracing.SpanFromContext(ctx).Context()))
	ctx = opentracing.ContextWithSpan(ctx, span)
	tryApplyCommonInfo(span)
	tag.SpanKindRPCClient.Set(span)
	tag.SpanKindConsumer.Set(span)
	tag.MessageBusDestination.Set(span, r.getFullIdentifier())
	return span, ctx
}

func tryApplyCommonInfo(span opentracing.Span) {
	tag.Component.Set(span, "azure-event-hubs-go")
	tryApplyNetworkInfo(span)
}

func tryApplyNetworkInfo(span opentracing.Span) {
	hostname, err := os.Hostname()
	if err == nil {
		tag.PeerHostname.Set(span, hostname)
	}
	_ = addIPAddressesToSpan(span)
}

func addIPAddressesToSpan(span opentracing.Span) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return err
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
	return nil
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

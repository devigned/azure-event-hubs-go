package tracing

import (
	"os"

	"github.com/opentracing/opentracing-go"
	tag "github.com/opentracing/opentracing-go/ext"
)

// ApplyComponentInfo applies eventhub library and network info to the span
func ApplyComponentInfo(span opentracing.Span) {
	tag.Component.Set(span, "github.com/Azure/azure-event-hubs-go")
	applyNetworkInfo(span)
}

func applyNetworkInfo(span opentracing.Span) {
	hostname, err := os.Hostname()
	if err == nil {
		tag.PeerHostname.Set(span, hostname)
	}
}

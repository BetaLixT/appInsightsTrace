package appinsightstrace

import "context"

// Implement this to extract w3c-trace information from the context, an example
// usage would be to build a middleware for incoming http calls to inject trace
// information into the context and create an implementation of ITraceExtractor
// to get the trace (using ctx.Value and with the correct key(s))
type ITraceExtractor interface {
  ExtractTraceInfo(
		ctx context.Context,
	) (ver, tid, pid, rid, flg string)
}

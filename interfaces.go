package appinsightstrace

import "context"

type ITraceExtractor interface {
  ExtractTraceInfo(
		ctx context.Context,
	) (ver, tid, pid, rid, flg string)
}

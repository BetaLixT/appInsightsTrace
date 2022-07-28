package appinsightstrace

import "context"

type DefaultTraceExtractor struct {
  
}

func (*DefaultTraceExtractor) ExtractTraceInfo(
		_ context.Context,
	) (ver, tid, pid, rid, flg string) {
	  return "", "", "", "", ""
	}

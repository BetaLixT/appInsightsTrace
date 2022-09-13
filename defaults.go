package appinsightstrace

import (
	"context"

	"github.com/soreing/trex"
)

type DefaultTraceExtractor struct {
}

const TRACE_INFO_KEY = "tinfo"

func (*DefaultTraceExtractor) ExtractTraceInfo(
	ctx context.Context,
) (ver, tid, pid, rid, flg string) {
	if tinfo, ok := ctx.Value(TRACE_INFO_KEY).(trex.TxModel); !ok {
		return "", "", "", "", ""
	} else {
		return tinfo.Ver, tinfo.Tid, tinfo.Pid, tinfo.Rid, tinfo.Flg
	}
}

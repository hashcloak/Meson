package epochtime

import (
	"context"
	"time"

	kpki "github.com/hashcloak/Meson/client/pkiclient"
	"github.com/hashcloak/Meson/katzenmint"
)

//! The duration of a katzenmint epoch. Should refer to katzenmint PKI.
var TestPeriod = katzenmint.HeightPeriod * time.Duration(katzenmint.EpochInterval)

func Now(client kpki.Client) (epoch uint64, ellapsed, till time.Duration, err error) {
	preparingEpoch, ellapsedHeight, err := client.GetEpoch(context.Background())
	epoch = preparingEpoch - 1
	ellapsed = katzenmint.SinceEpochStart(int64(ellapsedHeight))
	till = katzenmint.TillEpochFinish(int64(ellapsedHeight))
	return
}

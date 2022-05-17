package epochtime

import (
	"context"
	"time"

	kpki "github.com/hashcloak/Meson/client/pkiclient"
	"github.com/hashcloak/Meson/katzenmint"
)

//! The duration of a katzenmint epoch. Should refer to katzenmint PKI.
var TestPeriod = 10 * time.Second

//! Number of heights across an epoch. Should refer to katzenmint PKI.
var epochInterval = uint64(katzenmint.EpochInterval)

func Now(client kpki.Client) (epoch uint64, ellapsed, till time.Duration, err error) {
	preparingEpoch, ellapsedHeight, err := client.GetEpoch(context.Background())
	epoch = preparingEpoch - 1
	if ellapsedHeight > epochInterval {
		ellapsedHeight = epochInterval
	}
	ellapsed = time.Duration(uint64(TestPeriod) * ellapsedHeight / epochInterval)
	till = TestPeriod - ellapsed
	return
}

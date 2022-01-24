package epochtime

import (
	"context"
	"time"

	kpki "github.com/hashcloak/Meson/client/pkiclient"
)

//! The duration of a katzenmint epoch. Should refer to katzenmint PKI.
var TestPeriod = 10 * time.Second

//! Number of heights across an epoch. Should refer to katzenmint PKI.
var testEpochInterval uint64 = 10

func Now(client kpki.Client) (epoch uint64, ellapsed, till time.Duration, err error) {
	epoch, ellapsedHeight, err := client.GetEpoch(context.Background())
	if ellapsedHeight > testEpochInterval {
		ellapsedHeight = testEpochInterval
	}
	ellapsed = time.Duration(uint64(TestPeriod) * ellapsedHeight / testEpochInterval)
	till = TestPeriod - ellapsed
	// epoch should minus 1
	// TODO: find a better way for this
	epoch -= 1
	return
}

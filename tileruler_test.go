package tileruler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_hexStr2int(t *testing.T) {
	Convey("Convert hex format string to decimal", t, func() {
		hexDecs := map[string]int{
			"1":   1,
			"002": 2,
			"011": 17,
		}

		for hex, dec := range hexDecs {
			val, err := hexStr2int(hex)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, dec)
		}
	})

}

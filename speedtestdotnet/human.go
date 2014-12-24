package speedtestdotnet

import (
	"fmt"
)

var (
	bits        = 8
	kb          = uint64(1024)
	mb          = 1024 * kb
	gb          = 1024 * mb
	tb          = 1024 * gb
	pb          = 1024 * tb
	tooDamnFast = "Too fast to test"
)

func HumanSpeed(bps uint64) string {
	if bps > pb {
		return tooDamnFast
	} else if bps > tb {
		return fmt.Sprintf("%.02f Tb/s", float64(bps)/float64(tb))
	} else if bps > gb {
		return fmt.Sprintf("%.02f Gb/s", float64(bps)/float64(gb))
	} else if bps > mb {
		return fmt.Sprintf("%.02f Mb/s", float64(bps)/float64(mb))
	} else if bps > kb {
		return fmt.Sprintf("%.02f Kb/s", float64(bps)/float64(kb))
	}
	return fmt.Sprintf("%d bps", bps)
}

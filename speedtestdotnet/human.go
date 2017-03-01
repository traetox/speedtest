// The MIT License (MIT)

// Copyright (c) 2014, 2016 traetox

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package speedtestdotnet

import (
	"fmt"
)

const (
	bits        = 8
	kb          = 1024
	mb          = 1024 * kb
	gb          = 1024 * mb
	tb          = 1024 * gb
	pb          = 1024 * tb
	KB          = 1024 * bits
	MB          = 1024 * KB
	GB          = 1024 * MB
	TB          = 1024 * GB
	PB          = 1024 * TB
	tooDamnFast = "Too fast to test"
)

func HumanSpeed(bps uint64) string {
	if bps > pb {
		return tooDamnFast
	} else if bps > tb {
		return fmt.Sprintf("%.02f Tbits/s %.02f TBytes/s", float64(bps)/float64(tb), float64(bps)/float64(TB))
	} else if bps > gb {
		return fmt.Sprintf("%.02f Gbits/s %.02f GBytes/s", float64(bps)/float64(gb), float64(bps)/float64(GB))
	} else if bps > mb {
		return fmt.Sprintf("%.02f Mbit/s %.02f MBytes/s", float64(bps)/float64(mb), float64(bps)/float64(MB))
	} else if bps > kb {
		return fmt.Sprintf("%.02f Kbits/s  %.02f KBytes/s", float64(bps)/float64(kb), float64(bps)/float64(KB))
	}
	return fmt.Sprintf("%d bps", bps)
}

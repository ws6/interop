package interop

import (
	"encoding/binary"

	"os"
	"time"
)

var (
	WinTimeKindMask     = uint64(0X3FFFFFFFFFFFFFFF)
	WINDOWS_TICK        = uint64(1e7)
	SEC_SINCE_WIN_EPOCH = uint64(62135596800) //0001, Jan,1st, 12:00
)

type ExtractionMetrics struct {
	LaneNum     uint16
	TileNum     uint16
	Cycle       uint16
	Fwhm_A      float32
	Fwhm_C      float32
	Fwhm_G      float32
	Fwhm_T      float32
	Intensity_A uint16
	Intensity_C uint16
	Intensity_G uint16
	Intensity_T uint16
	CIF_TIME    uint64
}

type WinTime struct {
	DateTime uint64
}

type ExtractionInfo struct {
	Filename string
	Version  uint8
	SSize    uint8
	Metrics  []*ExtractionMetrics
	MaxCycle uint64
	err      error
}

//WinToUnixTimeStamp RTA windows timestamp to linux timestamp
func WinToUnixTimeStamp(ts uint64) uint64 {
	//first two bits are date "kind"; not used
	ts &= WinTimeKindMask
	seconds := (ts / WINDOWS_TICK)
	return (seconds - uint64(SEC_SINCE_WIN_EPOCH))
}

func GetTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

func (self *ExtractionInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.Filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	header, err := GetHeader(file)

	if err != nil {
		self.err = err
		return self.err
	}
	self.Version = header.Version
	self.SSize = header.SSize
	for {
		em := new(ExtractionMetrics)
		err = binary.Read(header.Buf, binary.LittleEndian, em)
		if err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			break
		}
		em.CIF_TIME = WinToUnixTimeStamp(em.CIF_TIME)
		self.Metrics = append(self.Metrics, em)
	}

	return self.err
}

//TODO interface to all Metrics; add General Stat Function instead of compute each time
func (self *ExtractionInfo) GetLaneMaxCycle() map[uint16]uint16 {
	laneMaxCycle := make(map[uint16]uint16)
	for _, v := range self.Metrics {
		if _, ok := laneMaxCycle[v.LaneNum]; !ok {
			laneMaxCycle[v.LaneNum] = v.Cycle
		}

		if v.Cycle > laneMaxCycle[v.LaneNum] {
			laneMaxCycle[v.LaneNum] = v.Cycle
		}
	}
	return laneMaxCycle
}

func (self *ExtractionInfo) GetMinCycle() uint16 {
	m := self.GetLaneMaxCycle()

	minCycle := uint16(0)
	i := 0
	for _, cycle := range m {
		i++
		if i == 1 {
			minCycle = cycle
			continue
		}
		if cycle < minCycle {
			minCycle = cycle
		}
	}

	return minCycle
}

func (self *ExtractionInfo) GetMaxCycle() uint16 {
	m := self.GetLaneMaxCycle()
	maxCycle := uint16(0)
	for _, cycle := range m {
		if cycle > maxCycle {
			maxCycle = cycle
		}
	}
	return maxCycle
}

func (self *ExtractionInfo) GetLatestCIFTime() int64 {
	latest := uint64(0)
	for _, v := range self.Metrics {
		if v.CIF_TIME > latest {
			latest = v.CIF_TIME
		}
	}
	return int64(latest)
}

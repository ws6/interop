package interop

import (
	"encoding/binary"
	"math"
	"os"
)

type ErrorMetrics struct {
	LaneNum         uint16
	TileNum         uint16
	Cycle           uint16
	ErrorRate       float32
	NumPerfectReads uint32
	Num_1_Error     uint32
	Num_2_Error     uint32
	Num_3_Error     uint32
	Num_4_Error     uint32
}

type ErrorInfo struct {
	Filename string
	Version  uint8
	SSize    uint8
	Metrics  []*ErrorMetrics
	err      error
}

func (self *ErrorInfo) Parse() error {
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
		em := new(ErrorMetrics)
		err = binary.Read(header.Buf, binary.LittleEndian, em)
		if err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			break
		}

		self.Metrics = append(self.Metrics, em)
	}
	return self.err
}

func (self *ErrorInfo) FilterByTileMap(tm *[]LaneTile) *ErrorInfo {
	ret := &ErrorInfo{
		Filename: self.Filename,
		Version:  self.Version,
		SSize:    self.SSize,
	}

	tmap := MakeLaneTileMap(tm)
	ret.Metrics = make([]*ErrorMetrics, 0)
	for _, t := range self.Metrics {
		if _, ok := tmap[t.LaneNum]; !ok {
			continue
		}
		if use, ok := tmap[t.LaneNum][t.TileNum]; !ok || !use {
			continue
		}
		//!!! only use ref
		ret.Metrics = append(ret.Metrics, t)
	}
	return nil
}

//GetAvgErrorRateByLane if cycleMap is nil, not to use
func (self *ErrorInfo) GetAvgErrorRateByLane(laneNum uint16, cycleMap *map[uint16]bool) float64 {
	sum := float64(0)
	cnt := 0
	for _, m := range self.Metrics {
		if m.LaneNum != laneNum {
			continue
		}
		if cycleMap != nil {
			dref := *cycleMap
			if _, ok := dref[m.Cycle]; !ok {
				continue
			}
		}

		cnt++
		sum += float64(m.ErrorRate)
	}
	if cnt == 0 {
		return float64(0.0)
	}
	return sum / float64(cnt)

}

func (self *ErrorInfo) GetStatErrorRateByLane(laneNum uint16, cycleMap *map[uint16]bool) (mean float64, stdv float64) {
	sum, devsum := float64(0), float64(0)
	cnt := 0
	for _, m := range self.Metrics {
		if m.LaneNum != laneNum {
			continue
		}
		if cycleMap != nil {
			dref := *cycleMap
			if _, ok := dref[m.Cycle]; !ok {
				continue
			}
		}

		cnt++
		sum += float64(m.ErrorRate)
	}
	if cnt == 0 {
		return
	}
	mean = sum / float64(cnt)

	for _, m := range self.Metrics {
		if m.LaneNum != laneNum {
			continue
		}
		if cycleMap != nil {
			dref := *cycleMap
			if _, ok := dref[m.Cycle]; !ok {
				continue
			}
		}

		v := float64(m.ErrorRate)
		b := mean - v
		devsum += (b * b)
	}
	stdv = math.Sqrt(devsum / float64(cnt))

	return

}

package interop

import (
	"encoding/binary"
	"fmt"

	"io"

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
type LTC3 struct {
	LaneNum uint16
	//	LaneNum1 uint8
	//	LaneNum2 uint8
	TileNum uint32
	//	T1    uint8
	//	T2    uint8
	//	T3    uint8
	//	T4    uint8
	Cycle uint16
}

type ExtractionMetricsV3 struct {
	LTC3
	Fwhm      []float32
	Intensity []uint16
}

type WinTime struct {
	DateTime uint64
}

type ExtractionInfo struct {
	Filename    string
	Filenames   []string // for V3
	Version     uint8
	SSize       uint8
	NumChannels uint8
	Metrics     []*ExtractionMetrics
	Metrics3    []*ExtractionMetricsV3
	MaxCycle    uint64
	err         error
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
func (self *ExtractionInfo) Parse3(filename string) error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	//	if err := binary.Read(file, binary.LittleEndian, &self.NumChannels); err != nil {
	if err := binary.Read(file, binary.LittleEndian, &self.Version); err != nil {
		return err
	}
	if self.Version != 3 {
		return fmt.Errorf("not RTA3 - %d", self.Version)
	}
	if err := binary.Read(file, binary.LittleEndian, &self.SSize); err != nil {
		return err
	}
	if err := binary.Read(file, binary.LittleEndian, &self.NumChannels); err != nil {
		return err
	}

	fmt.Println(`number of channels `, self.NumChannels, self.SSize, self.Version)
	//	return fmt.Errorf(`term`)
	//	self.NumChannels = 2
	for {
		em := new(ExtractionMetricsV3)
		//		if err := binary.Read(file, binary.LittleEndian, &em.LTC3); err != nil {
		if err := binary.Read(file, binary.LittleEndian, &em.LTC3); err != nil {

			if err == io.EOF {
				return nil
			}
			return err
		}
		//		fmt.Printf("%+v\n", em.LTC3)
		//TODO read FWHM

		//TODO read intensity
		for i := 0; i < int(self.NumChannels); i++ {
			var f float32
			if err := binary.Read(file, binary.LittleEndian, &f); err != nil {

				if err == io.EOF {
					return nil
				}
				return err
			}
			em.Fwhm = append(em.Fwhm, f)
		}

		for i := 0; i < int(self.NumChannels); i++ {
			var intensity uint16

			if err := binary.Read(file, binary.LittleEndian, &intensity); err != nil {

				if err == io.EOF {
					return nil
				}
				return err
			}
			em.Intensity = append(em.Intensity, intensity)
		}
		//		fmt.Printf("%+v\n", em.Intensity)
		self.Metrics3 = append(self.Metrics3, em)
	}

	return nil
}

//caller will open and close the file handler

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
		if v.LaneNum == 0 {
			continue
		}
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
func (self *ExtractionInfo) GetFirstCIFTime() int64 {
	first := uint64(time.Now().Unix())
	for _, v := range self.Metrics {
		if v.CIF_TIME < first {
			first = v.CIF_TIME
		}
	}
	return int64(first)
}

func (self *ExtractionInfo) FilterByTileMap(tm *[]LaneTile) *ExtractionInfo {
	ret := &ExtractionInfo{
		Filename: self.Filename,
		Version:  self.Version,
		SSize:    self.SSize,
		MaxCycle: self.MaxCycle,
	}

	tmap := MakeLaneTileMap(tm)
	ret.Metrics = make([]*ExtractionMetrics, 0)
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
	return ret
}

package interop

import (
	"encoding/binary"
	"math"
	"os"
	"sort"
)

type ReadCode struct {
	ReadNum,
	Phasing, PrePhasing, PercentAligned uint16
}

//1-based read number to code set
func ReadNumToCode(N uint16) ReadCode {
	return ReadCode{
		Phasing:        200 + (N-1)*2,
		PrePhasing:     201 + (N-1)*2,
		PercentAligned: 300 + N - 1,
	}
}

var (
	CLUSTER_DENSITY    = uint16(100)
	CLUSTER_DENSITY_PF = uint16(101)
	NUMBER_CLUSTER     = uint16(102)
	NUMBER_CLUSTER_PF  = uint16(103)

	PHASING_READ1     = uint16(200)
	PRE_PHASING_READ1 = uint16(201)
)

//TODO dynamic calc names
func TileCodeToName(code int) string {
	switch code {
	case 0:
		return "undefined"
	case 100:
		return "cluster_density"
	case 101:
		return "cluster_density_pf"
	case 102:
		return "num_clusters"
	case 103:
		return "num_clusters_pf"
		//phasing prephasing
	case 200:
		return "phasing_read1"
	case 201:
		return "prephasing_read1"
	case 202:
		return "phasing_read2"
	case 203:
		return "prephasing_read2"
	case 204:
		return "phasing_read3"
	case 205:
		return "prephasing_read3"
	case 206:
		return "phasing_read4"
	case 207:
		return "prephasing_read4"
	//percent aligned
	case 300:
		return "percent_aligned_read1"
	case 301:
		return "percent_aligned_read2"
	case 302:
		return "percent_aligned_read3"
	case 303:
		return "percent_aligned_read4"
	case 400:
		return "control_lane"

	}

	return TileCodeToName(0)
}

type TileMetrics struct {
	LaneNum     uint16
	TileNum     uint16
	MetricCode  uint16
	MetricValue float32
}

type TileInfo struct {
	Filename string
	Version  uint8
	SSize    uint8
	Metrics  []*TileMetrics
	err      error
}

func (self *TileInfo) Parse() error {
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
		em := new(TileMetrics)
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

func (self *TileInfo) FilterByTileMap(tm *[]LaneTile) *TileInfo {
	ret := &TileInfo{
		Filename: self.Filename,
		Version:  self.Version,
		SSize:    self.SSize,
	}
	tmap := MakeLaneTileMap(tm)
	ret.Metrics = make([]*TileMetrics, 0)
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

func (self *TileInfo) GetLanesSorted() []uint16 {
	lm := make(map[int]bool)
	for _, v := range self.Metrics {
		lm[int(v.LaneNum)] = true
	}
	keys := make([]int, len(lm))
	i := 0
	for k, _ := range lm {
		keys[i] = k
		i++
	}
	ret := make([]uint16, 0)

	sort.Ints(keys)
	for _, v := range keys {
		ret = append(ret, uint16(v))
	}
	return ret
}

func (self *TileInfo) CodeSumByLane() map[uint16]map[uint16]float64 {
	ret := make(map[uint16]map[uint16]float64)
	if self == nil {
		return ret
	}
	for _, cv := range self.Metrics {
		if _, ok := ret[cv.LaneNum]; !ok {
			ret[cv.LaneNum] = make(map[uint16]float64)
		}
		ret[cv.LaneNum][cv.MetricCode] += float64(cv.MetricValue)
	}
	return ret
}

func (self *TileInfo) CodeAvgByLane(laneNum, Code uint16) float64 {
	sum := float64(0.0)
	count := 0
	for _, cv := range self.Metrics {
		if cv.LaneNum != laneNum {
			continue
		}
		if cv.MetricCode != Code {
			continue
		}
		count++
		sum += float64(cv.MetricValue)
	}
	if count == 0 {
		return float64(0.0)
	}
	return sum / float64(count)
}

func GetTileStat(a *[]float64) (mean float64, stdev float64) {
	sum, devsum := float64(0.0), float64(0.0)
	arr := *a
	num := len(arr)
	if num == 0 {
		return
	}
	for _, v := range arr {
		sum += v
	}
	mean = sum / float64(num)
	for _, v := range arr {
		b := mean - v
		devsum += (b * b)
	}
	stdev = math.Sqrt(devsum / float64(num))
	return
}

func (self *TileInfo) GetAlignedStatByLanes(reads, lanes []uint16) (float64, float64) {

	er := make([]float64, 0) //error rates
	codes := make(map[uint16]bool)
	for _, r := range reads {
		//get percent aligned code
		c := ReadNumToCode(r).PercentAligned
		codes[c] = true
	}
	lnmap := make(map[uint16]bool)
	for _, l := range lanes {
		lnmap[l] = true
	}
	for _, cv := range self.Metrics {
		//not in lanes
		if ok, exists := lnmap[cv.LaneNum]; !exists || !ok {
			continue
		}
		//not in reads=>PercentAligned codes
		if ok, exists := codes[cv.MetricCode]; !exists || !ok {
			continue
		}
		er = append(er, float64(cv.MetricValue))
	}
	return GetTileStat(&er)
}

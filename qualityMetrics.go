package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

//common struct Lane,Tile and Cycle
var (
	QSCORE_UPPER = 100
)

type LTC struct {
	LaneNum uint16
	TileNum uint16
	Cycle   uint16
}

type QMetrics struct {
	LTC
	NumClusters [50]uint32 //first 50 clusters by cycle score Q1 through Q50 (uint32
}
type QbinConfig struct {
	LowerBound  []uint8
	UpperBound  []uint8
	ReMapScores []uint8
}

type QMetricsInfo struct {
	Filename   string
	Version    uint8
	SSize      uint8
	EnableQbin bool
	NumQscores uint8 //number of quality score bins, B
	QbinConfig QbinConfig
	Metrics    []*QMetrics
	err        error
}

func (self *QMetricsInfo) Error() string {
	if self.err != nil {
		if self.err.Error() == "EOF" {
			return ""
		}
		return self.err.Error()
	}
	return ""
}

func (self *QMetricsInfo) ParseQbinConfig(buffer *bufio.Reader) error {
	err := binary.Read(buffer, binary.LittleEndian, &self.NumQscores)
	if err != nil {
		self.err = err
		return err
	}
	self.QbinConfig.LowerBound = make([]uint8, self.NumQscores)
	err = binary.Read(buffer, binary.LittleEndian, &self.QbinConfig.LowerBound)
	if err != nil {
		self.err = err
		return err
	}
	self.QbinConfig.UpperBound = make([]uint8, self.NumQscores)
	err = binary.Read(buffer, binary.LittleEndian, &self.QbinConfig.UpperBound)
	if err != nil {
		self.err = err
		return err
	}

	self.QbinConfig.ReMapScores = make([]uint8, self.NumQscores)
	err = binary.Read(buffer, binary.LittleEndian, &self.QbinConfig.ReMapScores)
	if err != nil {
		self.err = err
		return err
	}
	return nil
}

func (self *QMetricsInfo) ParseQbin(buffer *bufio.Reader) error {
	if err := self.ParseQbinConfig(buffer); err != nil {
		return err
	}
	return self.ParseNonQbin(buffer)
}
func (self *QMetricsInfo) ParseNonQbin(buffer *bufio.Reader) error {
	if self.err != nil {
		return self.err
	}
	for {
		m := new(QMetrics)
		err := binary.Read(buffer, binary.LittleEndian, m)
		if err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			break
		}
		self.Metrics = append(self.Metrics, m)
	}
	return self.err
}

func boundCheck(arr []uint8, errTag string) error {
	for i, v := range arr {
		if v > 60 {
			return fmt.Errorf("%s: %d >60 at %d", errTag, v, i)
		}
	}
	return nil
}

func (self *QMetricsInfo) ValidateQbinConfig() error {
	if self.NumQscores > 50 {
		return fmt.Errorf("qbin szie:%d >50", self.NumQscores)
	}

	if err := boundCheck(self.QbinConfig.LowerBound, "lower bound"); err != nil {
		return err
	}
	if err := boundCheck(self.QbinConfig.UpperBound, "upper bound"); err != nil {
		return err
	}
	if err := boundCheck(self.QbinConfig.ReMapScores, "remap score"); err != nil {
		return err
	}
	return nil
}

func (self *QMetricsInfo) ParseVersion6(buffer *bufio.Reader) error {
	if self.err = self.ParseQbinConfig(buffer); self.err != nil {
		return self.err
	}
	fmt.Printf("%+v\n", self.Version)
	fmt.Printf("%+v\n", self.QbinConfig)
	if self.EnableQbin {
		if self.err = self.ValidateQbinConfig(); self.err != nil {
			fmt.Printf("%+v\n", self.Version)
			fmt.Printf("%+v\n", self.QbinConfig)
			return self.err
		}
	}

	ok := true
	n := uint32(0)
	for {
		if self.err != nil {
			if self.err.Error() == "EOF" {
				self.err = nil
			}
			break
		}
		m := new(QMetrics)
		if self.EnableQbin {
			if self.err = binary.Read(buffer, binary.LittleEndian, &m.LTC); self.err != nil {
				continue
			}
			ok = true
			for i := uint8(0); i < self.NumQscores; i++ {
				n = 0
				if self.err = binary.Read(buffer, binary.LittleEndian, &n); self.err != nil {
					ok = false
					break
				}
				m.NumClusters[self.QbinConfig.ReMapScores[i]] = n
			}
			if !ok {
				continue
			}

		} else {
			if self.err = binary.Read(buffer, binary.LittleEndian, m); self.err != nil {
				continue
			}
		}

		if self.err == nil {
			self.Metrics = append(self.Metrics, m)
		}

	}

	return self.err
}

func (self *QMetricsInfo) Parse() error {
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
	self.EnableQbin = false

	if self.Version == 4 {
		return self.ParseNonQbin(header.Buf)
	}

	if self.Version == 5 {
		var enableQbined uint8
		err := binary.Read(header.Buf, binary.LittleEndian, &enableQbined)
		if err != nil {
			self.err = err
			return self.err
		}
		if enableQbined == 1 {
			self.EnableQbin = true
			return self.ParseQbin(header.Buf)
		}
	}
	if self.Version == 6 {
		var enableQbined uint8
		if self.err = binary.Read(header.Buf, binary.LittleEndian, &enableQbined); self.err != nil {
			return self.err
		}

		if enableQbined == 1 {
			self.EnableQbin = true
			return self.ParseVersion6(header.Buf)
		}
	}
	return self.ParseNonQbin(header.Buf)
}

func (self *QMetricsInfo) FilterByTileMap(tm *[]LaneTile) *QMetricsInfo {
	ret := &QMetricsInfo{
		Filename:   self.Filename,
		Version:    self.Version,
		SSize:      self.SSize,
		EnableQbin: self.EnableQbin,
		NumQscores: self.NumQscores,
		QbinConfig: self.QbinConfig,
	}

	tmap := MakeLaneTileMap(tm)
	ret.Metrics = make([]*QMetrics, 0)
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

func (self *QMetricsInfo) GetLaneMaxCycle() map[uint16]uint16 {
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

//func (self *QMetricsInfo) GetStatByCycle(inlanes []uint16, inCycles []uint16) (mean float64, stdev float64) {

//	count := uint64(0)
//	qscoreTotal := float64(0)

//	filterLane := func() bool {

//		return false
//	}

//	for _, m := range self.Metrics {

//		qscoreTotal += float64( ) * float64(qval)
//		count += uint64(qscore)
//	}
//	mean = float64(0)
//	if count != 0 {
//		mean = qscoreTotal / float64(count)
//	}
//	stdevSum := float64(0)
//	//	for qval, qscore := range laneSum[laneNum] {
//	//		//TODO add filter
//	//		stdevSum += float64(qscore) * (math.Pow((mean - float64(qval)), float64(2)))
//	//	}

//	if count != 0 {
//		stdev = math.Sqrt(stdevSum / float64(count))
//	}
//	return
//}

//GetLaneSum return either filtered or unfiltered by cycleMap
func (self *QMetricsInfo) GetLaneSum(cycleMap *map[uint16]bool) map[uint16][]uint64 {
	ret := make(map[uint16][]uint64)
	for _, v := range self.Metrics {
		if cycleMap != nil {
			cm := *cycleMap
			if used, ok := cm[v.Cycle]; !ok || !used {
				continue
			}
		}
		if _, ok := ret[v.LaneNum]; !ok {
			ret[v.LaneNum] = make([]uint64, len(v.NumClusters))
		}
		//		ref := ret[v.LaneNum]
		for qval, qscore := range v.NumClusters {

			ret[v.LaneNum][qval] += uint64(qscore)
		}
	}
	return ret
}

func (self *QMetricsInfo) QscoreInCycle(qvalCutoff int, laneNum uint16, cycleMap map[uint16]bool) uint64 {
	count := uint64(0)
	for _, v := range self.Metrics {
		if v.LaneNum != laneNum {
			continue
		}
		if used, ok := cycleMap[v.Cycle]; !ok || !used {
			continue
		}
		for qval, qscore := range v.NumClusters {
			if (qval) >= qvalCutoff {
				count += uint64(qscore)
			}
		}
	}

	return count
}

func QscoreSumByLane(laneSum map[uint16][]uint64, laneNum uint16, qvalCutoff int) uint64 {
	count := uint64(0)
	if _, ok := laneSum[laneNum]; !ok {
		return count
	}
	for qval, qscore := range laneSum[laneNum] {

		if (qval + 1) >= qvalCutoff {
			count += qscore
		}
	}
	return count
}

func QvalToErrorRate(qval int) float64 {
	return math.Pow(float64(10), (float64(-1.0)*float64(qval))/float64(10))
}

//ExpectErrorRateStat return qval to error rate mean/stdev
func ExpectErrorRateStat(laneSum map[uint16][]uint64, laneNum uint16) (mean float64, stdev float64) {
	if _, ok := laneSum[laneNum]; !ok {
		return
	}
	count := uint64(0)
	errorSum := float64(0)
	for qval, qscore := range laneSum[laneNum] {
		errorRate := QvalToErrorRate(qval)
		errorSum += float64(qscore) * errorRate
		count += qscore
	}
	mean = float64(0)
	if count != 0 {
		mean = errorSum / float64(count)
	}
	stdevSum := float64(0)
	for qval, qscore := range laneSum[laneNum] {
		errorRate := QvalToErrorRate(qval)
		stdevSum += float64(qscore) * (math.Pow((mean - errorRate), float64(2)))
	}

	if count != 0 {
		stdev = math.Sqrt(stdevSum / float64(count))
	}
	return
}

//QscoreLaneStat return mean and stdev
func QscoreLaneStat(laneSum map[uint16][]uint64, laneNum uint16) (mean float64, stdev float64) {
	if _, ok := laneSum[laneNum]; !ok {
		return
	}
	count := uint64(0)
	qscoreTotal := float64(0)
	for qval, qscore := range laneSum[laneNum] {

		qscoreTotal += float64(qscore) * float64(qval)
		count += uint64(qscore)
	}
	mean = float64(0)
	if count != 0 {
		mean = qscoreTotal / float64(count)
	}
	stdevSum := float64(0)
	for qval, qscore := range laneSum[laneNum] {
		stdevSum += float64(qscore) * (math.Pow((mean - float64(qval)), float64(2)))
	}

	if count != 0 {
		stdev = math.Sqrt(stdevSum / float64(count))
	}
	return
}

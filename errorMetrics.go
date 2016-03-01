package interop

import (
	"encoding/binary"
	//	"fmt"
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

type TileErrorRate struct {
	TileNum       uint16    `json:"t"` //short json size
	ErrorRates    []float32 `json:"-"` //all cycles
	MeanErrorRate float32   `json:"v"`
}

type LaneErrorRate struct {
	LaneNum  uint16
	Surfaces [][][]*TileErrorRate //sorted [surface][swath]
}
type TileDimension struct {
	ValueName    string
	Surface      uint16
	Swath        uint16
	TilesInSwath uint16
	Lanes        []uint16
}

type FlowcellErrorRate struct {
	Dim            TileDimension
	Lanes          []*LaneErrorRate
	LaneNumToIndex []uint16 //Lookup for laneNum to index ae Lanes LaneIndex[Lane4]->1; maybe map but not ideal
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
func hi(in uint16) (uint16, uint16) {
	x := in
	i := uint16(0)
	for x != 0 {
		if x < 10 {
			break
		}
		i++
		x = x / 10

	}
	tens := uint16(1)
	for i != 0 {
		i--
		tens *= 10
	}
	remains := in - x*tens
	return x, remains
}

func GetTileDim(tileNum uint16) TileDimension {
	ret := TileDimension{}
	ret.Surface, ret.TilesInSwath = hi(tileNum)
	ret.Swath, ret.TilesInSwath = hi(ret.TilesInSwath)
	return ret
}

//TODO interface
func (self *ErrorInfo) GetDimMax() TileDimension {
	ret := TileDimension{}
	laneMap := make(map[uint16]bool)
	for _, m := range self.Metrics {
		dim := GetTileDim(m.TileNum)
		if dim.Surface > ret.Surface {
			ret.Surface = dim.Surface
		}
		if dim.Swath > ret.Swath {
			ret.Swath = dim.Swath
		}
		if dim.TilesInSwath > ret.TilesInSwath {
			ret.TilesInSwath = dim.TilesInSwath
		}
		if _, ok := laneMap[m.LaneNum]; !ok {
			laneMap[m.LaneNum] = true
		}
	}
	for ln, _ := range laneMap {
		ret.Lanes = append(ret.Lanes, ln)
	}
	return ret
}

func MeanStatFloat32(a *[]float32) (mean float32, stdev float32) {
	sum, devsum := float64(0.0), float64(0.0)
	arr := *a
	num := len(arr)
	if num == 0 {
		return
	}
	for _, v := range arr {
		sum += float64(v)
	}
	mean = float32(sum / float64(num))
	for _, v := range arr {
		b := mean - v
		devsum += float64(b * b)
	}
	stdev = float32(math.Sqrt(float64(devsum) / float64(num)))
	return
}

func (self *ErrorInfo) ErrorRateByTile() FlowcellErrorRate {
	ret := FlowcellErrorRate{}
	//init

	dim := self.GetDimMax()
	ret.Dim = dim
	ret.Dim.ValueName = "Error Rate"
	//	fmt.Printf("%+v\n", dim)
	ret.Lanes = make([]*LaneErrorRate, len(dim.Lanes))
	maxLenNum := uint16(0)
	for i, ln := range dim.Lanes {
		if ln > maxLenNum {
			maxLenNum = ln
		}
		ret.Lanes[i] = new(LaneErrorRate)
		ret.Lanes[i].LaneNum = ln

	}
	//create look up
	ret.LaneNumToIndex = make([]uint16, maxLenNum+1)
	for i, ln := range dim.Lanes {
		ret.LaneNumToIndex[ln] = uint16(i)
	}
	//init surface, swath
	for _, lr := range ret.Lanes {
		lr.Surfaces = make([][][]*TileErrorRate, dim.Surface)
		for surface := uint16(0); surface < dim.Surface; surface++ {
			lr.Surfaces[surface] = make([][]*TileErrorRate, dim.Swath)
			for swath := uint16(0); swath < dim.Swath; swath++ {
				lr.Surfaces[surface][swath] = make([]*TileErrorRate, dim.TilesInSwath)
				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					lr.Surfaces[surface][swath][swathTiles] = new(TileErrorRate)
				}
			}
		}
	}

	//load
	for _, m := range self.Metrics {

		tileDim := GetTileDim(m.TileNum)
		laneIndex := ret.LaneNumToIndex[m.LaneNum]
		surfaceIndex := tileDim.Surface - 1
		swathIndex := tileDim.Swath - 1
		swathTileIndex := tileDim.TilesInSwath - 1
		tile := ret.Lanes[laneIndex].Surfaces[surfaceIndex][swathIndex][swathTileIndex]
		tile.TileNum = m.TileNum
		tile.ErrorRates = append(tile.ErrorRates, m.ErrorRate)

	}
	//compute
	for _, lr := range ret.Lanes {

		for surface := uint16(0); surface < dim.Surface; surface++ {

			for swath := uint16(0); swath < dim.Swath; swath++ {

				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					tile := lr.Surfaces[surface][swath][swathTiles]
					tile.MeanErrorRate, _ = MeanStatFloat32(&tile.ErrorRates)

				}
			}
		}
	}

	return ret
}

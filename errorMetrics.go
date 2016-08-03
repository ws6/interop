package interop

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
)

var (
	BUBBLE_THRESH_RATE = float32(1.5)
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
	Cycle        uint16 //total cycles
	TilesInSwath uint16
	Lanes        []uint16
}

type FlowcellErrorRate struct {
	Dim              TileDimension
	Lanes            []*LaneErrorRate
	LaneNumToIndex   []uint16 //Lookup for laneNum to index ae Lanes LaneIndex[Lane4]->1; maybe map but not ideal
	TotalValidCycles int
}

type BubbleSum struct {
	BubbledTiles       int
	TotalBubbles       int // across all tiles
	TotalValidCycles   int
	BubbleRate         float32
	MeanInBubbledTiles float32
	MeanInAllTiles     float32
	FlowcellErrorRate
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
		if m.Cycle > ret.Cycle {
			ret.Cycle = m.Cycle
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

//GetWhiteErrorRate only given empty of ErrorRateMetrics
func GetWhiteErrorRate(dim *TileDimension) FlowcellErrorRate {
	ret := FlowcellErrorRate{}
	//init

	ret.Dim = *dim
	ret.Dim.ValueName = "Blank Tile Map"
	fmt.Printf("%+v\n", ret.Dim)
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
	numPad := int(math.Log10(float64(dim.TilesInSwath)) + 1)
	formatter := fmt.Sprintf("%%d%%d%%0%dd", numPad)
	for _, lr := range ret.Lanes {
		lr.Surfaces = make([][][]*TileErrorRate, dim.Surface)
		for surface := uint16(0); surface < dim.Surface; surface++ {
			lr.Surfaces[surface] = make([][]*TileErrorRate, dim.Swath)
			for swath := uint16(0); swath < dim.Swath; swath++ {
				lr.Surfaces[surface][swath] = make([]*TileErrorRate, dim.TilesInSwath)
				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					te := new(TileErrorRate)
					//TODO add tile number here
					tileStr := fmt.Sprintf(formatter, surface+1, swath+1, swathTiles+1)
					tn, err := strconv.Atoi(tileStr)
					if err != nil {
						fmt.Println(err.Error())
					}
					te.TileNum = uint16(tn)
					lr.Surfaces[surface][swath][swathTiles] = te
				}
			}
		}
	}

	return ret
}

func (self *ErrorInfo) ErrorRateByTile(_dim *TileDimension) FlowcellErrorRate {
	ret := FlowcellErrorRate{}
	//init

	dim := self.GetDimMax()
	if _dim != nil { //overwrite from the given; Voyager patch
		//		dim.Lanes = _dim.Lanes
		dim.Surface = _dim.Surface
		dim.Swath = _dim.Swath
		dim.TilesInSwath = _dim.TilesInSwath
		//		dim.Cycle = _dim.Cycle
	}
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
	numPad := int(math.Log10(float64(dim.TilesInSwath)) + 1)
	formatter := fmt.Sprintf("%%d%%d%%0%dd", numPad)
	for _, lr := range ret.Lanes {
		lr.Surfaces = make([][][]*TileErrorRate, dim.Surface)
		for surface := uint16(0); surface < dim.Surface; surface++ {
			lr.Surfaces[surface] = make([][]*TileErrorRate, dim.Swath)
			for swath := uint16(0); swath < dim.Swath; swath++ {
				lr.Surfaces[surface][swath] = make([]*TileErrorRate, dim.TilesInSwath)
				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					te := new(TileErrorRate)
					//TODO add tile number here
					tileStr := fmt.Sprintf(formatter, surface+1, swath+1, swathTiles+1)
					tn, err := strconv.Atoi(tileStr)
					if err != nil {
						fmt.Println(err.Error())
					}
					te.TileNum = uint16(tn)
					lr.Surfaces[surface][swath][swathTiles] = te
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

//BubbleCount return valid cycles
func (self *TileErrorRate) BubbleCount(excludeCycles map[uint16]bool) int {
	//no valid data
	validCycle := 0
	if self.TileNum == uint16(0) {
		return validCycle
	}
	sz := len(self.ErrorRates)
	for cycle, cur := range self.ErrorRates {
		if excludeCycles != nil {
			if _, ok := excludeCycles[uint16(cycle+1)]; ok {
				continue
			}
		}

		//skip first cycle
		if cycle == 0 {
			continue
		}
		//skip last two cycles
		if (cycle) >= sz-2 {
			continue
		}
		//skip current error rate zero
		if cur == float32(0) {
			continue
		}
		pre := self.ErrorRates[cycle-1]
		//skip pre no error rate
		if pre == float32(0) {
			continue
		}
		next := self.ErrorRates[cycle+1]
		//skip next no error rate
		if next == float32(0) {
			continue
		}
		validCycle++

		if cur >= BUBBLE_THRESH_RATE*(pre+next) {
			//			fmt.Println(self.TileNum, cycle-1, cycle, cycle+1, pre, cur, next, BUBBLE_THRESH_RATE*(pre+next))
			self.MeanErrorRate += float32(1) // the correct name should be num bubbles
		}

	}
	return validCycle
}

//count all cycles before filter. let controller to do filtering
func (self *ErrorInfo) BubbleCounter(excludeCycles map[uint16]bool) FlowcellErrorRate {

	ret := FlowcellErrorRate{}
	//init

	dim := self.GetDimMax()
	ret.Dim = dim
	ret.Dim.ValueName = "Bubble Count"
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
	//init surface, swath, cycles
	for _, lr := range ret.Lanes {
		lr.Surfaces = make([][][]*TileErrorRate, dim.Surface)
		for surface := uint16(0); surface < dim.Surface; surface++ {
			lr.Surfaces[surface] = make([][]*TileErrorRate, dim.Swath)
			for swath := uint16(0); swath < dim.Swath; swath++ {
				lr.Surfaces[surface][swath] = make([]*TileErrorRate, dim.TilesInSwath)
				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					er := new(TileErrorRate)
					er.ErrorRates = make([]float32, dim.Cycle)
					lr.Surfaces[surface][swath][swathTiles] = er
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
		tile.ErrorRates[m.Cycle-1] = m.ErrorRate

	}
	//compute
	for _, lr := range ret.Lanes {

		for surface := uint16(0); surface < dim.Surface; surface++ {

			for swath := uint16(0); swath < dim.Swath; swath++ {

				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					tile := lr.Surfaces[surface][swath][swathTiles]
					validCycles := tile.BubbleCount(excludeCycles)
					ret.TotalValidCycles += validCycles
					//					if tile.MeanErrorRate > float32(0) {
					//						fmt.Println(lr.LaneNum, surface, swath, swathTiles, tile.TileNum)
					//					}
				}
			}
		}
	}

	return ret
}

func (self *FlowcellErrorRate) GetBubbleSum(_dim *TileDimension) BubbleSum {
	ret := BubbleSum{
		FlowcellErrorRate: *self,
		TotalValidCycles:  self.TotalValidCycles,
	}
	dim := self.Dim
	fmt.Printf("%+v\n", dim)
	//	if _dim != nil { //overwrite from the given; Voyager patch
	//		//		dim.Lanes = _dim.Lanes
	//		dim.Surface = _dim.Surface
	//		dim.Swath = _dim.Swath
	//		dim.TilesInSwath = _dim.TilesInSwath
	//		//		dim.Cycle = _dim.Cycle
	//	}
	//	fmt.Printf("%+v\n", dim)
	//TODO parsing dim
	for _, lr := range self.Lanes {

		for surface := uint16(0); surface < dim.Surface; surface++ {

			for swath := uint16(0); swath < dim.Swath; swath++ {

				for swathTiles := uint16(0); swathTiles < dim.TilesInSwath; swathTiles++ {
					tile := lr.Surfaces[surface][swath][swathTiles]
					//remove small fraction of float hack, maybe not have
					if tile.MeanErrorRate > float32(0) {
						ret.BubbledTiles++
						ret.TotalBubbles += int(tile.MeanErrorRate)
					}
				}
			}
		}
	}
	if ret.TotalBubbles != 0 {
		//TODO re-define it using total valide cycles
		ret.BubbleRate = float32(ret.TotalBubbles) / float32(ret.TotalValidCycles)
	}
	if ret.BubbledTiles != 0 {
		ret.MeanInBubbledTiles = float32(ret.TotalBubbles) / float32(ret.BubbledTiles)
	}
	//	if _dim != nil {
	//		ret.FlowcellErrorRate.Dim.TilesInSwath = _dim.TilesInSwath
	//		//		ret.FlowcellErrorRate.Dim.Surface = _dim.Surface
	//		//		ret.FlowcellErrorRate.Dim.Swath = _dim.Swath
	//	}

	return ret
}

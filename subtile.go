package interop

import (
	"fmt"
	"sort"
)

var (
	SUBTYPE_PF          = "PF"
	SUBTYPE_DENSITY_RAW = "Density Raw"
	SUBTYPE_DENSITY_PF  = "Density PF"

	SUBTYPE_CLUSTER_RAW = "Cluster Raw"
	SUBTYPE_CLUSTER_PF  = "Cluster PF"

	SUBTYPE_FWHM = "FWHM"
	A_CHAL       = 0
	G_CHAL       = 1
	C_CHAL       = 2
	T_CHAL       = 3
)

type BinStat struct {
	numbers  []float64 `json:"-"`
	BinValue uint16    //Nx, Ny
	*BoxWhiskerStat
}

//combine these two
type SubtileInfo struct {
	PFInfo   *PFMetricsInfo
	FwhmInfo *FwhmMetricsInfo
	SubtileLaneStat
}

type BinStatMap struct {
	XBinStat map[uint16]map[uint16]*BinStat
	YBinStat map[uint16]map[uint16]*BinStat
}

type SubtileLaneStat struct {
	PF                  *BinStatMap
	DensityRaw          *BinStatMap //need transform from BinArea
	DensityPF           *BinStatMap
	ClusterRaw          *BinStatMap
	ClusterPF           *BinStatMap
	FWHM_Channels       []*BinStatMap // A G C T
	FWHMAll_Channel_All *BinStatMap   // with channels and all
}

type convert func(m *PFSubTileMetrics, Ny, x, y uint16) float64
type convertFwhm func(m *FwhmSubTileMetrics, Ny, x, y uint16) float64
type postProcess func(*SubtileInfo, *BinStatMap) error

func (self *SubtileInfo) Validate() error {
	if self.PFInfo.BinArea == 0 {
		return fmt.Errorf("self.PFInfo.BinArea is zero")
	}

	if self.FwhmInfo == nil {
		return nil
	}
	if self.PFInfo.NumX != uint16(self.FwhmInfo.NumX) {
		return fmt.Errorf("PF X (%d) is no same as FWHM's (%d)", self.PFInfo.NumX, self.FwhmInfo.NumX)
	}
	if self.PFInfo.NumY != uint16(self.FwhmInfo.NumY) {
		return fmt.Errorf("PF Y (%d) is no same as FWHM's (%d)", self.PFInfo.NumY, self.FwhmInfo.NumY)
	}
	return nil
}

func (self *SubtileInfo) GetMetricsFiltered(getter convert, postfn postProcess, filter func(lane, tile uint16) bool) error {
	if err := self.Validate(); err != nil {
		return err
	}

	XBinStat := make(map[uint16]map[uint16]*BinStat) // lane ->Xbin ->stat
	YBinStat := make(map[uint16]map[uint16]*BinStat) //lane->Ybin->stat

	for _, m := range self.PFInfo.Metrics {
		if !filter(m.LaneNum, m.TileNum) {
			continue
		}
		if _, ok := XBinStat[m.LaneNum]; !ok {
			XBinStat[m.LaneNum] = make(map[uint16]*BinStat)
		}
		if _, ok := YBinStat[m.LaneNum]; !ok {
			YBinStat[m.LaneNum] = make(map[uint16]*BinStat)
		}
		for x := uint16(0); x < self.PFInfo.NumX; x++ {
			if _, ok := XBinStat[m.LaneNum][x]; !ok {
				XBinStat[m.LaneNum][x] = new(BinStat)
				XBinStat[m.LaneNum][x].BinValue = x
			}
			for y := uint16(0); y < self.PFInfo.NumY; y++ {
				if _, ok := YBinStat[m.LaneNum][y]; !ok {
					YBinStat[m.LaneNum][y] = new(BinStat)
					YBinStat[m.LaneNum][y].BinValue = y
				}

				val := getter(m, self.PFInfo.NumY, x, y)
				XBinStat[m.LaneNum][x].numbers = append(XBinStat[m.LaneNum][x].numbers, val)
				YBinStat[m.LaneNum][y].numbers = append(YBinStat[m.LaneNum][y].numbers, val)
				//TODO push to binX and binY
			}
		}
	}
	//compute box whisker stat
	computeBoxStat := func(m map[uint16]map[uint16]*BinStat) {
		for _, binMap := range m {
			for _, _stat := range binMap {
				_stat.BoxWhiskerStat = new(BoxWhiskerStat)
				_stat.BoxWhiskerStat.GetFloat64(&_stat.numbers)
			}
		}
	}

	computeBoxStat(XBinStat)
	computeBoxStat(YBinStat)

	return postfn(self, &BinStatMap{XBinStat, YBinStat})
}

func (self *SubtileInfo) GetPFSubTileMetrics(getter convert, postfn postProcess) error {
	if err := self.Validate(); err != nil {
		return err
	}

	XBinStat := make(map[uint16]map[uint16]*BinStat) // lane ->Xbin ->stat
	YBinStat := make(map[uint16]map[uint16]*BinStat) //lane->Ybin->stat

	for _, m := range self.PFInfo.Metrics {
		if _, ok := XBinStat[m.LaneNum]; !ok {
			XBinStat[m.LaneNum] = make(map[uint16]*BinStat)
		}
		if _, ok := YBinStat[m.LaneNum]; !ok {
			YBinStat[m.LaneNum] = make(map[uint16]*BinStat)
		}
		for x := uint16(0); x < self.PFInfo.NumX; x++ {
			if _, ok := XBinStat[m.LaneNum][x]; !ok {
				XBinStat[m.LaneNum][x] = new(BinStat)
				XBinStat[m.LaneNum][x].BinValue = x
			}
			for y := uint16(0); y < self.PFInfo.NumY; y++ {
				if _, ok := YBinStat[m.LaneNum][y]; !ok {
					YBinStat[m.LaneNum][y] = new(BinStat)
					YBinStat[m.LaneNum][y].BinValue = y
				}

				val := getter(m, self.PFInfo.NumY, x, y)
				XBinStat[m.LaneNum][x].numbers = append(XBinStat[m.LaneNum][x].numbers, val)
				YBinStat[m.LaneNum][y].numbers = append(YBinStat[m.LaneNum][y].numbers, val)
				//TODO push to binX and binY
			}
		}
	}
	//compute box whisker stat
	computeBoxStat := func(m map[uint16]map[uint16]*BinStat) {
		for _, binMap := range m {
			for _, _stat := range binMap {
				_stat.BoxWhiskerStat = new(BoxWhiskerStat)
				_stat.BoxWhiskerStat.GetFloat64(&_stat.numbers)
			}
		}
	}

	computeBoxStat(XBinStat)
	computeBoxStat(YBinStat)

	return postfn(self, &BinStatMap{XBinStat, YBinStat})
}

func (self *SubtileInfo) GetClusterRawMetrics() error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		return float64(m.RawCluster[self.PFInfo.NumY*x+y])
	}
	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.ClusterRaw = binMap
		return nil
	}
	return self.GetPFSubTileMetrics(getter, postfn)
}

func (self *SubtileInfo) GetClusterPFMetrics() error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		return float64(m.PFCluster[self.PFInfo.NumY*x+y])
	}
	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.ClusterPF = binMap
		return nil
	}
	return self.GetPFSubTileMetrics(getter, postfn)
}

//GetPFSubTileMetricsFiltered

func (self *SubtileInfo) GetPFMetricsFiltered(filter func(lane, tile uint16) bool) error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		return float64(m.PFCluster[self.PFInfo.NumY*x+y])
	}
	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.ClusterPF = binMap
		return nil
	}
	return self.GetMetricsFiltered(getter, postfn, filter)
}

//TODO add ClusterRaw filtered function
func (self *SubtileInfo) GetRawClusterMetricsFiltered(filter func(lane, tile uint16) bool) error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		return float64(m.RawCluster[self.PFInfo.NumY*x+y])
	}
	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.ClusterRaw = binMap
		return nil
	}
	return self.GetMetricsFiltered(getter, postfn, filter)
}

func (self *SubtileInfo) GetPctPF_Filtered(filter func(lane, tile uint16) bool) error {
	//load ClusterPF
	//load clusterRaw
	//compute pct pf
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		if m.RawCluster[self.PFInfo.NumY*x+y] == 0 {
			return 0
		}
		return float64(100.*m.PFCluster[self.PFInfo.NumY*x+y]) / float64(m.RawCluster[self.PFInfo.NumY*x+y])

	}

	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.PF = binMap
		return nil
	}
	return self.GetMetricsFiltered(getter, postfn, filter)
	//	return self.GetPFSubTileMetrics(getter, postfn)
}

func (self *SubtileInfo) GetPFMetrics() error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		if m.RawCluster[self.PFInfo.NumY*x+y] == 0 {
			return 0
		}
		return float64(m.PFCluster[self.PFInfo.NumY*x+y]) / float64(m.RawCluster[self.PFInfo.NumY*x+y])

	}

	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.PF = binMap
		return nil
	}
	return self.GetPFSubTileMetrics(getter, postfn)
}

func (self *SubtileInfo) GetDenstiyMetrics() error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		if m.RawCluster[self.PFInfo.NumY*x+y] == 0 {
			return 0
		}

		density := float64(m.RawCluster[self.PFInfo.NumY*x+y]) / float64(self.PFInfo.BinArea)
		return density / 1000. //convert to k/mm2
	}
	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.DensityRaw = binMap
		return nil
	}
	return self.GetPFSubTileMetrics(getter, postfn)
}

func (self *SubtileInfo) GetDenstiyPFMetrics() error {
	getter := func(m *PFSubTileMetrics, Ny, x, y uint16) float64 {
		if m.RawCluster[self.PFInfo.NumY*x+y] == 0 {
			return 0
		}

		density := float64(m.PFCluster[self.PFInfo.NumY*x+y]) / float64(self.PFInfo.BinArea)
		return density / 1000. //convert to k/mm2
	}
	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		self.DensityPF = binMap
		return nil
	}
	return self.GetPFSubTileMetrics(getter, postfn)
}

func (self *SubtileInfo) GetFWHMSubTileMetrics(getter convertFwhm, postfn postProcess) error {
	if err := self.Validate(); err != nil {
		return err
	}

	XBinStat := make(map[uint16]map[uint16]*BinStat) // lane ->Xbin ->stat
	YBinStat := make(map[uint16]map[uint16]*BinStat) //lane->Ybin->stat

	for _, m := range self.FwhmInfo.Metrics {
		if _, ok := XBinStat[m.LaneNum]; !ok {
			XBinStat[m.LaneNum] = make(map[uint16]*BinStat)
		}
		if _, ok := YBinStat[m.LaneNum]; !ok {
			YBinStat[m.LaneNum] = make(map[uint16]*BinStat)
		}

		for x := uint16(0); x < uint16(self.FwhmInfo.NumX); x++ {
			if _, ok := XBinStat[m.LaneNum][x]; !ok {
				XBinStat[m.LaneNum][x] = new(BinStat)
				XBinStat[m.LaneNum][x].BinValue = x
			}
			for y := uint16(0); y < uint16(self.FwhmInfo.NumY); y++ {
				if _, ok := YBinStat[m.LaneNum][y]; !ok {
					YBinStat[m.LaneNum][y] = new(BinStat)
					YBinStat[m.LaneNum][y].BinValue = y
				}

				//				val := float64(m.RawCluster[self.PFInfo.NumY*x+y])
				//				val := getter(m, uint16(self.FwhmInfo.NumY), x, y)
				val := getter(m, uint16(self.FwhmInfo.NumX), y, x)

				XBinStat[m.LaneNum][x].numbers = append(XBinStat[m.LaneNum][x].numbers, val)
				YBinStat[m.LaneNum][y].numbers = append(YBinStat[m.LaneNum][y].numbers, val)
				//TODO push to binX and binY
			}
		}
		//		}
	}
	//compute box whisker stat
	computeBoxStat := func(m map[uint16]map[uint16]*BinStat) {
		for _, binMap := range m {
			for _, _stat := range binMap {
				_stat.BoxWhiskerStat = new(BoxWhiskerStat)
				_stat.BoxWhiskerStat.GetFloat64(&_stat.numbers)
			}
		}
	}

	computeBoxStat(XBinStat)
	computeBoxStat(YBinStat)

	return postfn(self, &BinStatMap{XBinStat, YBinStat})
}

func (self *SubtileInfo) GetFwhmMetricsByChannel(channelIndex uint16) error {
	//channelIndex 0, 1, 3,4 ->A G C T
	if channelIndex >= uint16(self.FwhmInfo.NumChannels) {
		return nil
		//		return fmt.Errorf("channelIndex(%d) is greater than 3", channelIndex)
	}
	getter := func(m *FwhmSubTileMetrics, Ny, x, y uint16) float64 {
		metrics := m.Channels[channelIndex]
		return float64(metrics.Fwhm[Ny*x+y])
	}

	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {
		//		//!!! make once only
		//		if len(self.FWHM_Channels) == 0 {
		//			self.FWHM_Channels = make([]*BinStatMap, 4)
		//		}
		self.FWHM_Channels[channelIndex] = binMap
		return nil
	}
	return self.GetFWHMSubTileMetrics(getter, postfn)
}

func (self *SubtileInfo) GetFwhmMetricsByAllChannel() error {
	getter := func(m *FwhmSubTileMetrics, Ny, x, y uint16) float64 {

		total := float64(0)
		for i := 0; i < int(self.FwhmInfo.NumChannels); i++ {
			metrics := m.Channels[i]
			total += float64(metrics.Fwhm[Ny*x+y])
		}
		if self.FwhmInfo.NumChannels == 0 {
			return 0.
		}
		//!!! not always divided by 4, need check number of Channels used
		return total / float64(self.FwhmInfo.NumChannels)
	}

	postfn := func(self *SubtileInfo, binMap *BinStatMap) error {

		self.FWHMAll_Channel_All = binMap
		return nil
	}
	return self.GetFWHMSubTileMetrics(getter, postfn)
}

//GetFwhmMetrics compute four possible channels and all by average
func (self *SubtileInfo) GetFwhmMetrics() error {
	if self.FwhmInfo == nil {
		return nil
	}
	allChannels := []int{A_CHAL, G_CHAL, C_CHAL, T_CHAL}
	for _, c := range allChannels {
		if err := self.GetFwhmMetricsByChannel(uint16(c)); err != nil {
			return err
		}
	}
	//todo make all
	return self.GetFwhmMetricsByAllChannel()
}

type builder func() error

func (self *SubtileInfo) MakeBoxStat() error {
	//make four channels before started
	if len(self.FWHM_Channels) == 0 {
		self.FWHM_Channels = make([]*BinStatMap, self.FwhmInfo.NumChannels)
	}

	builders := []builder{
		//		self.GetFwhmMetrics,
		self.GetClusterRawMetrics,
		self.GetClusterPFMetrics,
		self.GetPFMetrics,
		self.GetDenstiyMetrics,
		self.GetDenstiyPFMetrics,
		self.GetFwhmMetricsByAllChannel,
	}

	allChannels := []int{A_CHAL, G_CHAL, C_CHAL, T_CHAL}
	for _, c := range allChannels {
		myFn := func(channel int) builder {
			return func() error {
				return self.GetFwhmMetricsByChannel(uint16(channel))
			}
		}(c)
		builders = append(builders, myFn)

	}

	errc := make(chan error)
	result := make(chan int)
	for _, fn := range builders {
		go func(b builder) {
			if err := b(); err != nil {
				errc <- err
			}
			result <- 1
		}(fn)
	}
	var err error
	for i := 0; i < len(builders); i++ {
		select {
		case <-result:
		case err = <-errc:
		}
	}
	return err
}

type BinLaneStat struct {
	LaneNum    uint16
	BinBoxStat []*BinStat
}

type BinStatJson struct {
	XBinStat []*BinLaneStat
	YBinStat []*BinLaneStat
	//	YBinStat map[string]map[string]*BinStat
}

type SubtileLaneStatJson struct {
	PF               *BinStatJson
	DensityRaw       *BinStatJson //need transform from BinArea
	DensityPF        *BinStatJson
	ClusterRaw       *BinStatJson
	ClusterPF        *BinStatJson
	FWHM_Channels    []*BinStatJson // A G C T
	FWHM_Channel_All *BinStatJson   // with channels and all
	FWHM_SIZE_LIMIT  int64
	FWHW_EXCEED      bool
}

func (self *SubtileLaneStat) ToJson() *SubtileLaneStatJson {
	ret := new(SubtileLaneStatJson)
	ret.PF = self.PF.ToJson()
	ret.DensityRaw = self.DensityRaw.ToJson()
	ret.DensityPF = self.DensityPF.ToJson()
	ret.ClusterRaw = self.ClusterRaw.ToJson()
	ret.ClusterPF = self.ClusterPF.ToJson()
	for _, v := range self.FWHM_Channels {
		ret.FWHM_Channels = append(ret.FWHM_Channels, v.ToJson())
	}
	ret.FWHM_Channel_All = self.FWHMAll_Channel_All.ToJson()
	return ret
}

func (self *BinStatMap) ToJson() *BinStatJson {
	if self == nil {
		return nil
	}
	ret := new(BinStatJson)
	numberLanes := len(self.XBinStat)
	ret.XBinStat = make([]*BinLaneStat, numberLanes)

	lanesX := []int{}
	for laneNum, _ := range self.XBinStat {
		lanesX = append(lanesX, int(laneNum))
	}
	sort.Ints(lanesX)
	for i, laneInt := range lanesX {
		//	for laneNum, binMap := range self.XBinStat {
		laneNum := uint16(laneInt)
		binMap := self.XBinStat[laneNum]
		ret.XBinStat[i] = new(BinLaneStat)
		ret.XBinStat[i].LaneNum = laneNum
		numberBins := len(binMap)
		ret.XBinStat[i].BinBoxStat = make([]*BinStat, numberBins)
		bins := []int{}
		for binVal, _ := range binMap {
			bins = append(bins, int(binVal))
		}
		sort.Ints(bins)
		for j, binVal := range bins {
			//		for _, stat := range binMap {
			stat := binMap[uint16(binVal)]
			ret.XBinStat[i].BinBoxStat[j] = stat

		}

	}

	numberLanes = len(self.YBinStat)
	ret.YBinStat = make([]*BinLaneStat, numberLanes)

	lanesY := []int{}
	for laneNum, _ := range self.YBinStat {
		lanesY = append(lanesY, int(laneNum))
	}
	sort.Ints(lanesY)
	for i, laneInt := range lanesY {
		//	for laneNum, binMap := range self.XBinStat {
		laneNum := uint16(laneInt)
		binMap := self.YBinStat[laneNum]
		ret.YBinStat[i] = new(BinLaneStat)
		ret.YBinStat[i].LaneNum = laneNum
		numberBins := len(binMap)
		ret.YBinStat[i].BinBoxStat = make([]*BinStat, numberBins)
		bins := []int{}
		for binVal, _ := range binMap {
			bins = append(bins, int(binVal))
		}
		sort.Ints(bins)
		for j, binVal := range bins {
			//		for _, stat := range binMap {
			stat := binMap[uint16(binVal)]
			ret.YBinStat[i].BinBoxStat[j] = stat
		}
	}
	return ret
}

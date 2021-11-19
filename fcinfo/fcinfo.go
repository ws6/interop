package fcinfo

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const INTEROP_EXPIRED_HOURS = 24

var (
	NORUNPARAM = fmt.Errorf("No RunParamters.xml")
	VOYAGER    = "Voyager"
	FIRE_FLY   = "Firefly"
)

func IsNoRunParamErr(err error) bool {
	return err == NORUNPARAM
}

//in general what shall be common attributes for a flowcell
type Flowcell struct {
	FlowcellBarcode      string
	RunId                string
	FCPosition           string
	RunNumber            string
	Indexed              string
	ReadLength           string
	MachineName          string
	ApplicationName      string
	ApplicationVersion   string
	FpgaVersion          string
	RtaVersion           string
	RunParamOutputFolder string
	Description          string
	Location             string
	RunStartDate         string
	InstrumentType       string
	Chemistry            string
	Cycles               int
	RecipePath           string
}

type RunFolder struct {
	RunFolder            string
	RunInfoFileName      string
	RunParameterFileName string
	InterOpExists        bool
	DataExists           bool
	Err                  error
}

type RunParamSetUp struct {
	ApplicationName string `xml:"ApplicationName"`

	ApplicationVersion string `xml:"ApplicationVersion"`
	FCPosition         string `xml:"FCPosition"`
	//Hiseq,HiSeqX Sepcific
	RunParamOutPutFolder   string `xml:"OutputFolder"`
	FPGAVersion            string `xml:"FPGAVersion"`
	RTAVersion             string `xml:"RTAVersion"`
	ChemistryVersion       string `xml:"ChemistryVersion"`
	RecipeFragmentVTileMap string `xml:"RecipeFragmentVTileMap"`
}

type RunParams struct {
	//TODO add RecipeFragmentVersion
	ApplicationName       string
	ApplicationVersion    string
	FPGVersion            string
	RTAVersion            string
	RunParamOutPutFolder  string
	InstrumentType        string
	Chemistry             string
	RecipeFragmentVersion string
	FCPosition            string
	RunStartDate          string

	PlannedRead1Cycles      string
	PlannedRead2Cycles      string
	PlannedIndex1ReadCycles string
	PlannedIndex2ReadCycles string
	RecipePath              string
}

type RunParameters struct {
	Setup RunParamSetUp `xml:"Setup"`
	//FCPosition

	//Miseq,NextSeq,P11
	RunParamOutPutFolder string `xml:"OutputFolder"`
	FPGAVersion          string `xml:"FPGAVersion"`
	RTAVersion           string `xml:"RTAVersion"`
	Chemistry            string `xml:"Chemistry"`
	//Nova Seq
	Application string `xml:"Application"`
	//alter for NextSeq 1k/2k
	ApplicationName      string `xml:"ApplicationName"`
	ApplicationVersion   string `xml:"ApplicationVersion"`
	FlowCellSerialNumber string `xml:"FlowCellSerialNumber"`

	RecipePath string `xml:"RecipePath"`
}

type RunParamsPre struct {
	XMLName xml.Name `xml:"RunParameters"`

	RunParameters `xml:"RunParameters"`
}
type RunParamsVoyager struct {
	XMLName              xml.Name `xml:"RunParameters"`
	Application          string   `xml:"Application"`
	ApplicationVersion   string   `xml:"ApplicationVersion"`
	RunParamOutPutFolder string   `xml:"OutputFolder"`
	Autocenter           string   `xml:"Autocenter"`
	RunStartDate         string   `xml:"RunStartDate"`
	InstrumentName       string   `xml:"InstrumentName"`
	IsRehyb              string   `xml:"IsRehyb"`
	RecipeFilePath       string   `xml:"RecipeFilePath"`
	RunId                string   `xml:"RunId"`
	//	RunStartDate         string   `xml:"RunStartDate"`
	//	FPGAVersion            string   `xml:"FPGAVersion"`
	//	RTAVersion             string `xml:"RTAVersion"`
	//	ChemistryVersion       string `xml:"ChemistryVersion"`
	//	RecipeFragmentVTileMap string `xml:"RecipeFragmentVTileMap"`

}

type RunParamsFireflyProto2 struct {
	XMLName              xml.Name `xml:"RunParameters"`
	ApplicationName      string   `xml:"ApplicationName"`
	ApplicationVersion   string   `xml:"ApplicationVersion"`
	RunParamOutPutFolder string   `xml:"OutputFolderPath"`
}

//!!! For run info

//self close only use attributes of each <Read />
type RunInfoReads struct {
	// <Read Number="1" NumCycles="126" IsIndexedRead="N" />
	Number        int    `xml:"Number,attr"`
	NumCycles     int    `xml:"NumCycles,attr"`
	IsIndexedRead string `xml:"IsIndexedRead,attr"` //"N" or "Y"
	FirstCycle    int    `xml:"FirstCycle,attr"`
	LastCycle     int    `xml:"LastCycle,attr"`
	//	 FirstCycle="1" LastCycle="101"
}

//<FlowcellLayout LaneCount="1" SurfaceCount="2" SwathCount="1" TileCount="14" />
type RunInfoFlowcellLayout struct {
	LaneCount    int `xml:"LaneCount,attr"`
	SurfaceCount int `xml:"SurfaceCount,attr"`
	SwathCount   int `xml:"SwathCount,attr"`
	TileCount    int `xml:"TileCount,attr"`
}

//<Run Id="140203_M00805_0281_000000000-A7K65" Number="280">
type RunInfoRun struct {
	RunId           string                `xml:"Id,attr"`
	RunNumber       string                `xml:"Number,attr"`
	Date            string                `xml:"Date"`
	Instrument      string                `xml:"Instrument"`
	FlowcellBarcode string                `xml:"Flowcell"`
	FlowcellLayout  RunInfoFlowcellLayout `xml:"FlowcellLayout"`
	Reads           []RunInfoReads        `xml:"Reads>Read"`
}

type RunInfo struct {
	XMLName xml.Name `xml:"RunInfo"`
	Run     RunInfoRun
}

func ParseRunInfoXML(runInfoString string) (runInfo *RunInfo, err error) {
	runInfo = new(RunInfo)
	err = xml.Unmarshal([]byte(runInfoString), runInfo)
	return
}

func ParseRunParamsXMLPre(runParamString string) (runParamPre *RunParamsPre, err error) {
	runParamPre = new(RunParamsPre)
	err = xml.Unmarshal([]byte(runParamString), runParamPre)
	return
}

func ParseRunParamsXMLVoyager(runParamString string) (runParamPre *RunParamsVoyager, err error) {
	runParamPre = new(RunParamsVoyager)
	err = xml.Unmarshal([]byte(runParamString), runParamPre)
	return
}

//RunParamsFireflyProto2
func ParseRunParamsXMLFireflyProto2(runParamString string) (runParamPre *RunParamsFireflyProto2, err error) {
	runParamPre = new(RunParamsFireflyProto2)
	err = xml.Unmarshal([]byte(runParamString), runParamPre)
	return
}

func parseRunParamVoyager(runParamsPre *RunParamsVoyager) (runParam *RunParams, err error) {
	runParam = new(RunParams)

	runParam.ApplicationName = runParamsPre.Application

	runParam.ApplicationVersion = runParamsPre.ApplicationVersion
	runParam.InstrumentType = VOYAGER
	runParam.Chemistry = runParamsPre.RecipeFilePath
	runParam.FPGVersion = `undefined`
	runParam.RTAVersion = `undefined`
	runParam.RunParamOutPutFolder = runParamsPre.RunParamOutPutFolder
	runParam.RecipeFragmentVersion = `undefined`
	runParam.RunStartDate = runParamsPre.RunStartDate // overwrite from runInfo.xml's
	//guestimate the position
	sp := strings.Split(runParamsPre.RunId, "_")
	if len(sp) > 0 {
		positionAndBarcode := sp[len(sp)-1]
		if len(positionAndBarcode) > 0 {
			firstChar := positionAndBarcode[0]
			if firstChar == 'A' || firstChar == 'B' {
				runParam.FCPosition = string(firstChar)
			}
		}
	}
	return
}

func parseRunParamFireflyProto2(runParamsPre *RunParamsFireflyProto2) (runParam *RunParams, err error) {
	runParam = new(RunParams)

	runParam.ApplicationName = runParamsPre.ApplicationName
	runParam.ApplicationVersion = runParamsPre.ApplicationVersion
	runParam.InstrumentType = FIRE_FLY
	runParam.FPGVersion = `undefined`
	runParam.RTAVersion = `undefined`
	runParam.RunParamOutPutFolder = runParamsPre.RunParamOutPutFolder
	runParam.RecipeFragmentVersion = `undefined`
	//TODO get it from runInfo.xml
	//	runParam.RunStartDate = runParamsPre.RunStartDate // overwrite from runInfo.xml's

	return
}

func parseRunParamHiSeq(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam = new(RunParams)

	runParam.ApplicationName = runParamsPre.RunParameters.Setup.ApplicationName
	runParam.ApplicationVersion = runParamsPre.RunParameters.Setup.ApplicationVersion
	runParam.InstrumentType = "HiSeq"
	runParam.Chemistry = runParamsPre.RunParameters.Setup.ChemistryVersion
	runParam.FPGVersion = runParamsPre.RunParameters.Setup.FPGAVersion
	runParam.RTAVersion = runParamsPre.RunParameters.Setup.RTAVersion
	runParam.RunParamOutPutFolder = runParamsPre.RunParameters.Setup.RunParamOutPutFolder
	runParam.RecipeFragmentVersion = runParamsPre.RunParameters.Setup.RecipeFragmentVTileMap
	runParam.FCPosition = runParamsPre.RunParameters.Setup.FCPosition
	return
}

func parseRunParamHiSeqX(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam, err = parseRunParamHiSeq(runParamsPre)
	if err != nil {
		return
	}
	runParam.InstrumentType = "HiSeqX"
	return
}
func parseRunParamMiSeq(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam = new(RunParams)

	runParam.ApplicationName = runParamsPre.RunParameters.Setup.ApplicationName
	if runParam.ApplicationName == "" {
		runParam.ApplicationName = runParamsPre.ApplicationName
	}
	runParam.ApplicationVersion = runParamsPre.RunParameters.Setup.ApplicationVersion
	if runParam.ApplicationVersion == "" {
		runParam.ApplicationVersion = runParamsPre.ApplicationVersion
	}
	runParam.InstrumentType = "MiSeq"
	runParam.Chemistry = runParamsPre.RunParameters.Chemistry
	runParam.FPGVersion = runParamsPre.RunParameters.FPGAVersion
	runParam.RTAVersion = runParamsPre.RunParameters.RTAVersion
	runParam.RunParamOutPutFolder = runParamsPre.RunParameters.RunParamOutPutFolder
	return
}

//parseRunParamFirefly
func parseRunParamFirefly(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam = new(RunParams)

	runParam.ApplicationName = runParamsPre.RunParameters.Setup.ApplicationName
	runParam.ApplicationVersion = runParamsPre.RunParameters.Setup.ApplicationVersion
	runParam.InstrumentType = "Firefly"
	runParam.Chemistry = runParamsPre.RunParameters.Chemistry
	runParam.FPGVersion = runParamsPre.RunParameters.FPGAVersion
	runParam.RTAVersion = runParamsPre.RunParameters.RTAVersion
	runParam.RunParamOutPutFolder = runParamsPre.RunParameters.RunParamOutPutFolder
	runParam.RecipePath = runParamsPre.RunParameters.RecipePath

	return
}

//
func parseRunParamMerlion(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam = new(RunParams)

	runParam.ApplicationName = runParamsPre.RunParameters.Setup.ApplicationName
	runParam.ApplicationVersion = runParamsPre.RunParameters.Setup.ApplicationVersion
	runParam.InstrumentType = "Merlion"
	runParam.Chemistry = runParamsPre.RunParameters.Chemistry
	runParam.FPGVersion = runParamsPre.RunParameters.FPGAVersion
	runParam.RTAVersion = runParamsPre.RunParameters.RTAVersion
	runParam.RunParamOutPutFolder = runParamsPre.RunParameters.RunParamOutPutFolder
	return
}

//MiniSeq Control Software
func parseRunParamMiniSeq(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {

	info, err := parseRunParamMerlion(runParamsPre)
	if err != nil {
		return nil, err
	}
	info.InstrumentType = "MiniSeq"
	return info, nil
}

//Avatar Test Software

func parseRunParamAvatar(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {

	info, err := parseRunParamMerlion(runParamsPre)
	if err != nil {
		return nil, err
	}
	info.InstrumentType = "Avatar"
	return info, nil
}

func parseRunParamNextSeq(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam, err = parseRunParamMiSeq(runParamsPre)
	if err != nil {
		return
	}
	runParam.InstrumentType = "NextSeq"
	if strings.Contains(runParam.ApplicationName, "NextSeq 1000/2000") ||
		strings.Contains(runParam.ApplicationName, "Vega") {
		runParam.InstrumentType = `Vega`

	}
	return
}

func parseRunParamP11(runParamsPre *RunParamsPre) (runParam *RunParams, err error) {
	runParam, err = parseRunParamMiSeq(runParamsPre)
	if err != nil {
		return
	}
	runParam.InstrumentType = "P11"
	return
}

func ParseRunParamsXML(runParamString string) (*RunParams, error) {
	runParamPre, err := ParseRunParamsXMLPre(runParamString)

	if err != nil {
		return nil, err
	}

	//	ApplicationName := runParamPre.SetUp.ApplicationName
	ApplicationName := runParamPre.Setup.ApplicationName

	if ApplicationName == "" {
		ApplicationName = runParamPre.ApplicationName
	}

	matched, _ := regexp.MatchString("(?i)HiSeq X", ApplicationName)
	if matched {
		return parseRunParamHiSeqX(runParamPre)
	}
	matched, _ = regexp.MatchString("(?i)HiSeq", ApplicationName)
	if matched {
		return parseRunParamHiSeq(runParamPre)
	}

	matched, _ = regexp.MatchString("(?i)Merlion", ApplicationName)
	if matched {
		return parseRunParamMerlion(runParamPre)
	}
	matched, _ = regexp.MatchString("(?i)MiniSeq", ApplicationName)
	if matched {
		return parseRunParamMiniSeq(runParamPre)
	}
	//parseRunParamAvatar
	matched, _ = regexp.MatchString("(?i)Avatar", ApplicationName)
	if matched {
		return parseRunParamAvatar(runParamPre)
	}
	matched, _ = regexp.MatchString("(?i)MiSeq", ApplicationName)
	if matched {
		return parseRunParamMiSeq(runParamPre)
	}

	matched, _ = regexp.MatchString("(?i)Firefly", ApplicationName)
	if matched {
		firefly, _ := parseRunParamFirefly(runParamPre)

		return firefly, nil
	}

	matched, _ = regexp.MatchString("(?i)NextSeq", ApplicationName)
	if matched {
		return parseRunParamNextSeq(runParamPre)
	}

	matched, _ = regexp.MatchString("(?i)Project 11", ApplicationName)
	if matched {
		return parseRunParamP11(runParamPre)
	}

	//	matched, _ = regexp.MatchString("(?i)Voyager", ApplicationName)
	//	if matched {
	//		return parseRunParamVoyager(runParamPre)
	//	}
	//	fmt.Printf("%+v\n", runParamPre)
	//TODO parse out Voyager
	runParamPreVoyager, err := ParseRunParamsXMLVoyager(runParamString)
	if err == nil {
		matched, _ = regexp.MatchString("(?i)Voyager", runParamPreVoyager.Application)
		if matched {
			return parseRunParamVoyager(runParamPreVoyager)
		}
		//NovaSeq
		matched, _ = regexp.MatchString("(?i)NovaSeq", runParamPreVoyager.Application)
		if matched {
			return parseRunParamVoyager(runParamPreVoyager)
		}
		//		fmt.Printf("%+v\n", runParamPreVoyager)
	}
	//TODO parse firefly proto2
	//ParseRunParamsXMLFireflyProto2
	runParamPreFireflyProto2, err := ParseRunParamsXMLFireflyProto2(runParamString)
	if err == nil {
		matched, _ = regexp.MatchString("(?i)Firefly", runParamPreFireflyProto2.ApplicationName)
		if matched {
			return parseRunParamFireflyProto2(runParamPreFireflyProto2)
		}
		matchedISeq, _ := regexp.MatchString("(?i)iSeq", runParamPreFireflyProto2.ApplicationName)
		if matchedISeq {
			return parseRunParamFireflyProto2(runParamPreFireflyProto2)
		}

	}

	//iSeq Control Software

	return nil, errors.New("Can not find a properate parser for Application " + ApplicationName)
}

func calcCycles(runInfo *RunInfo) int {
	ret := 0
	ga2Cycle := 0
	for _, r := range runInfo.Run.Reads {

		ret += r.NumCycles
		ga2Cycle = r.LastCycle
	}
	if ret != 0 {
		return ret
	}
	return ga2Cycle
}

func GetCycles(runInfo *RunInfo) int {
	return calcCycles(runInfo)
}
func (runInfo *RunInfo) GetNumNonIndexReads() int {
	total := 0
	for _, r := range runInfo.Run.Reads {
		if r.IsIndexedRead != "Y" {
			total++
		}
	}
	return total
}
func isIndexed(runInfo *RunInfo) bool {
	for _, r := range runInfo.Run.Reads {
		if r.IsIndexedRead == "Y" {
			return true
		}
	}
	return false
}
func isIndexedString(runInfo *RunInfo) string {
	if isIndexed(runInfo) {
		return "yes"
	}
	return "no"
}

func (runInfo *RunInfo) GetNumReads() int {
	return len(runInfo.Run.Reads)
}

func (runInfo *RunInfo) GetNumCycles() int {
	ret := 0
	for _, r := range runInfo.Run.Reads {
		ret += int(r.NumCycles)
	}
	//for compatable with GAIIX RunInfo.xml
	if ret == 0 {
		for _, r := range runInfo.Run.Reads {
			ret = r.LastCycle
		}
	}
	return ret
}

func (runInfo *RunInfo) GetCyclesMapByRead(readNum int) map[uint16]bool {
	ret := make(map[uint16]bool)
	start, end := 0, 0
	for i := 0; i < readNum; i++ {
		if i >= len(runInfo.Run.Reads) {
			break
		}
		end += runInfo.Run.Reads[i].NumCycles
		start = end - runInfo.Run.Reads[i].NumCycles + 1
		//for GAIIX
		if runInfo.Run.Reads[i].NumCycles == 0 {
			end = runInfo.Run.Reads[i].LastCycle
			start = runInfo.Run.Reads[i].FirstCycle
		}
	}
	//Remove first and last cycle
	for i := start; i <= end; i++ {
		ret[uint16(i)] = true
	}
	return ret
}

//todo get first cycle map R1 && R2
func (runInfo *RunInfo) GetFirstCyclesMapByRead(readNum int) map[uint16]bool {
	ret := make(map[uint16]bool)
	start, end := 0, 0
	for i := 0; i < readNum; i++ {
		if i >= len(runInfo.Run.Reads) {
			break
		}
		end += runInfo.Run.Reads[i].NumCycles
		start = end - runInfo.Run.Reads[i].NumCycles + 1
		//for GAIIX
		if runInfo.Run.Reads[i].NumCycles == 0 {
			end = runInfo.Run.Reads[i].LastCycle
			start = runInfo.Run.Reads[i].FirstCycle
		}

	}
	ret[uint16(start)] = true
	return ret
}

func (runInfo *RunInfo) GetLastCyclesMapByRead(readNum int) map[uint16]bool {
	ret := make(map[uint16]bool)
	start, end := 0, 0
	_ = start
	for i := 0; i < readNum; i++ {
		if i >= len(runInfo.Run.Reads) {
			break
		}
		end += runInfo.Run.Reads[i].NumCycles
		start = end - runInfo.Run.Reads[i].NumCycles + 1
		//for GAIIX
		if runInfo.Run.Reads[i].NumCycles == 0 {
			end = runInfo.Run.Reads[i].LastCycle
			start = runInfo.Run.Reads[i].FirstCycle
		}

	}
	ret[uint16(end)] = true
	return ret
}

//Todo get last N cycle map R1 && R2
func (runInfo *RunInfo) GetLastNCyclesMapByRead(readNum, n int) map[uint16]bool {
	ret := make(map[uint16]bool)
	start, end := 0, 0
	for i := 0; i < readNum; i++ {
		if i >= len(runInfo.Run.Reads) {
			break
		}
		end += runInfo.Run.Reads[i].NumCycles
		start = end - runInfo.Run.Reads[i].NumCycles + 1
		//for GAIIX
		if runInfo.Run.Reads[i].NumCycles == 0 {
			end = runInfo.Run.Reads[i].LastCycle
			start = runInfo.Run.Reads[i].FirstCycle
		}

	}

	cycleStart := start
	if n > (end - start + 1) {
		cycleStart = start
	}

	for i := cycleStart; i <= end; i++ {
		ret[uint16(i)] = true
	}
	return ret
}

func (runInfo *RunInfo) GetFirstLastCyclesByRead(readNum int) []uint16 {
	ret := make([]uint16, 2)
	start, end := 0, 0
	for i := 0; i < readNum; i++ {
		if i >= len(runInfo.Run.Reads) {
			break
		}
		end += runInfo.Run.Reads[i].NumCycles
		start = end - runInfo.Run.Reads[i].NumCycles + 1
		//for GAIIX
		if runInfo.Run.Reads[i].NumCycles == 0 {
			end = runInfo.Run.Reads[i].LastCycle
			start = runInfo.Run.Reads[i].FirstCycle
		}

	}
	ret[0] = uint16(start)
	ret[1] = uint16(end)
	return ret
}

func calcReadLengthString(runInfo *RunInfo) string {
	ret := []string{}
	for _, r := range runInfo.Run.Reads {
		if r.NumCycles != 0 {
			ret = append(ret, strconv.Itoa(r.NumCycles))
			continue
		}
		ret = append(ret, strconv.Itoa(r.LastCycle-r.FirstCycle+1))

	}
	return strings.Join(ret, ",")
}

//GetRunType for venkman
func GetRunType(runInfo *RunInfo) string {
	return calcReadLengthString(runInfo)
}

func ParseFlowcell(runInfoString, runParamString string) (ret *Flowcell, err error) {
	ret = new(Flowcell)
	runParams, err := ParseRunParamsXML(runParamString)
	if err != nil {
		return
	}
	runInfo, err := ParseRunInfoXML(runInfoString)
	if err != nil {
		return
	}

	ret.RecipePath = runParams.RecipePath

	ret.ApplicationName = runParams.ApplicationName
	ret.ApplicationVersion = runParams.ApplicationVersion
	ret.Chemistry = runParams.Chemistry
	ret.Cycles = calcCycles(runInfo) //TODO calculate cycles
	ret.FlowcellBarcode = runInfo.Run.FlowcellBarcode
	ret.FpgaVersion = runParams.FPGVersion
	ret.Indexed = isIndexedString(runInfo) //TODO calculate is indexed
	ret.InstrumentType = runParams.InstrumentType
	ret.RtaVersion = runParams.RTAVersion
	ret.MachineName = runInfo.Run.Instrument

	ret.ReadLength = calcReadLengthString(runInfo) //TODO calc this
	ret.RunId = runInfo.Run.RunId
	if ret.MachineName == "" {
		sp := strings.Split(ret.RunId, "_")
		if len(sp) >= 2 {
			ret.MachineName = sp[1]
		}
	}
	//	ret.RunInfoXML = runInfoString
	//	ret.RunParamsXML = runParamString
	ret.RunStartDate = runInfo.Run.Date
	if runParams.InstrumentType == VOYAGER {
		ret.RunStartDate = runParams.RunStartDate
	}
	ret.RunNumber = runInfo.Run.RunNumber
	ret.RunParamOutputFolder = runParams.RunParamOutPutFolder
	ret.FCPosition = runParams.FCPosition
	//TODO convert runParams and runInfo into model.Flowcell
	// adding addtional server/location information?
	return ret, nil
}

func makePattern(dir, pat string) string {
	return fmt.Sprintf("%s", path.Join(dir, pat))
}
func ExistsOnePattern(files []string, pat string) (*string, error) {
	//case insentitive
	lowerPat := strings.ToLower(pat)

	for _, f := range files {
		lowerFfileName := filepath.Base(strings.ToLower(f))
		if lowerPat == lowerFfileName {
			return &f, nil
		}
	}
	return nil, nil
}

type PatternFound struct {
	Pat   string
	Found bool
	Name  string
}

func _CheckFlowcellRunFolder(dirIn string) (*RunFolder, error) {
	pats := []*PatternFound{
		&PatternFound{Pat: `data`},
		&PatternFound{Pat: `interop`},
		&PatternFound{Pat: `runinfo.xml`},
		&PatternFound{Pat: `runparameters.xml`},
	}

	find := func(pat string) string {
		for _, p := range pats {
			if pat == p.Pat {
				return p.Name
			}
		}
		return ""
	}
	ret := new(RunFolder)
	dir, err := filepath.Abs(dirIn)

	if err != nil {
		return nil, fmt.Errorf("filepath.Abs(dirIn) error : %s", dirIn)
	}
	ret.RunFolder = dir

	relFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range relFiles {
		//TODO hit found
		for _, p := range pats {
			if strings.ToLower(f.Name()) == p.Pat {
				p.Found = true
				p.Name = f.Name()
			}
		}
	}

	for _, p := range pats {
		if !p.Found {
			if p.Pat != `runparameters.xml` {
				return nil, fmt.Errorf(`not found %s`, p.Pat)
			}
			ret.Err = NORUNPARAM
		}
	}
	runInfoname := find(`runinfo.xml`)
	if runInfoname != "" {
		ret.RunInfoFileName = filepath.Join(dir, runInfoname)
	}
	ret.DataExists = true
	ret.InterOpExists = true
	runparamsname := find(`runparameters.xml`)
	if runparamsname != "" {
		ret.RunParameterFileName = filepath.Join(dir, runparamsname)
	}
	return ret, nil
}

func CheckFlowcellRunFolder(dirIn string) (*RunFolder, error) {
	ret := new(RunFolder)

	dir, err := filepath.Abs(dirIn)

	if err != nil {
		return nil, fmt.Errorf("filepath.Abs(dirIn) error : %s", dirIn)
	}

	//fast error return
	dataFolderJoin := filepath.Join(dir, "Data")
	interOpFolderJoin := filepath.Join(dir, "InterOp")

	if _, err := os.Stat(dataFolderJoin); err != nil {
		return nil, fmt.Errorf("read Data/ folder err:%s", err.Error())
	}

	if _, err := os.Stat(interOpFolderJoin); err != nil {
		return nil, fmt.Errorf("read InterOp/ folder err:%s", err.Error())
	}

	relFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir error:%s", err.Error())
	}

	files := []string{}

	for _, f := range relFiles {
		files = append(files, filepath.Join(dir, f.Name()))
	}

	ret.RunFolder = dir

	runInfo, err := ExistsOnePattern(files, "runInfo.xml")
	if err != nil {
		return nil, fmt.Errorf("RunInfo.xml error:%s", err.Error())
	}

	if runInfo == nil {
		return nil, fmt.Errorf("RunInfo.xml missing")
	}

	ret.RunInfoFileName = *runInfo

	dataFolder, err := ExistsOnePattern(files, "Data")
	if err != nil || dataFolder == nil {
		return nil, fmt.Errorf("Data folder missing")
	}

	ret.DataExists = true

	interOpFolder, err := ExistsOnePattern(files, "InterOp")
	if err != nil || interOpFolder == nil {
		return nil, fmt.Errorf("InterOp folder missing")
	}

	ret.InterOpExists = true

	runParameters, err := ExistsOnePattern(files, "runParameters.xml")
	if err != nil {
		return nil, fmt.Errorf("runParameters.xml error:%s", err.Error())
	}
	if runParameters == nil {
		ret.Err = NORUNPARAM //tolerate such error
		return ret, nil
	}

	ret.RunParameterFileName = *runParameters

	return ret, nil
}

func ParseGAFlowcell(runInfoString string) (ret *Flowcell, err error) {
	ret = new(Flowcell)

	runInfo, err := ParseRunInfoXML(runInfoString)
	if err != nil {
		return
	}
	ret.Description = "GA Family;no RunParameters.xml"
	ret.ApplicationName = "GA?" //TODO parse out from runInfo.xml
	//	ret.ApplicationVersion = runParams.ApplicationVersion
	//	ret.Chemistry = runParams.Chemistry
	ret.Cycles = calcCycles(runInfo) //TODO calculate cycles
	ret.FlowcellBarcode = runInfo.Run.FlowcellBarcode
	if ret.FlowcellBarcode == "" {
		ret.FlowcellBarcode = runInfo.Run.RunId
	}
	//	ret.FpgaVersion = runParams.FPGVersion
	ret.Indexed = isIndexedString(runInfo) //TODO calculate is indexed
	ret.InstrumentType = "GA?"
	//	ret.RtaVersion = runParams.RTAVersion
	ret.MachineName = runInfo.Run.Instrument
	ret.ReadLength = calcReadLengthString(runInfo) //TODO calc this
	ret.RunId = runInfo.Run.RunId
	//	ret.RunInfoXML = runInfoString
	//	ret.RunParamsXML = runParamString
	ret.RunStartDate = runInfo.Run.Date
	ret.RunNumber = runInfo.Run.RunNumber
	//	ret.RunParamOutputFolder = runParams.RunParamOutPutFolder
	//	ret.FCPosition = runParams.FCPosition
	//TODO convert runParams and runInfo into model.Flowcell
	// adding addtional server/location information?
	runIdInfo := strings.Split(runInfo.Run.RunId, "_")
	if len(runIdInfo) != 4 {
		return ret, nil
	}
	ret.RunStartDate = runIdInfo[0]
	instrument := runIdInfo[1]
	ret.RunNumber = runIdInfo[2]
	ret.FlowcellBarcode = runIdInfo[3]

	inst := strings.Split(instrument, "-")
	if len(inst) == 2 {
		ret.InstrumentType = inst[0]
		//		ret.ApplicationName = ret.InstrumentType
		//		ret.ApplicationVersion = ret.InstrumentType
	}
	return ret, nil

}

func _ParseFlowcellRunFolder(funFolder string, notSkipMissingRunParam bool) (ret *Flowcell, err error) {
	runfoldInfo, err := CheckFlowcellRunFolder(funFolder)
	if err != nil {
		return nil, err
	}
	runInfoByte, _err := ioutil.ReadFile(runfoldInfo.RunInfoFileName)
	if _err != nil {
		return nil, _err
	}
	if runfoldInfo.Err == NORUNPARAM {
		//reject parsing for later copied coming in
		//		return nil, NORUNPARAM
		//TODO get option from configuration to decide either fail or retry
		//For voyager team
		if notSkipMissingRunParam {
			return nil, NORUNPARAM
		}
		//for firefly team
		return ParseGAFlowcell(string(runInfoByte))
	}
	runParamsByte, err := ioutil.ReadFile(runfoldInfo.RunParameterFileName)
	if err != nil {
		return
	}
	return ParseFlowcell(string(runInfoByte), string(runParamsByte))
}

//ParseFlowcellRunInfo for path overwrite
func _ParseFlowcellRunInfo(runfoldInfo *RunFolder) (ret *Flowcell, err error) {

	runInfoByte, err := ioutil.ReadFile(runfoldInfo.RunInfoFileName)
	if err != nil {
		return
	}
	if runfoldInfo.Err == NORUNPARAM {
		return ParseGAFlowcell(string(runInfoByte))
	}
	runParamsByte, err := ioutil.ReadFile(runfoldInfo.RunParameterFileName)
	if err != nil {
		return
	}
	return ParseFlowcell(string(runInfoByte), string(runParamsByte))
}

func ParseFlowcellRunFolder(runFolder string, bearNoRunParam bool) (ret *Flowcell, err error) {
	ret, err = _ParseFlowcellRunFolder(runFolder, bearNoRunParam)
	if ret != nil {
		if len(ret.RunStartDate) > 10 {
			ret.RunStartDate = ret.RunStartDate[0:10]
		}
		if ret.FlowcellBarcode == "" {
			ret.FlowcellBarcode = ret.RunId
		}
		if ret.Location == "" {
			ret.Location = runFolder
		}
	}
	return ret, err
}

func ParseFlowcellRunInfo(runfoldInfo *RunFolder) (ret *Flowcell, err error) {
	ret, err = _ParseFlowcellRunInfo(runfoldInfo)
	if ret != nil {
		if len(ret.RunStartDate) > 10 {
			ret.RunStartDate = ret.RunStartDate[0:10]
		}
		if ret.FlowcellBarcode == "" {
			ret.FlowcellBarcode = ret.RunId
		}
	}
	return ret, err
}

func (self *RunInfo) CycleInRead(cycle int) *RunInfoReads {
	readStart := 0
	readEnd := 0
	for _, r := range self.Run.Reads {
		readEnd += r.NumCycles
		if cycle > readStart && cycle <= readEnd {
			return &r
		}
		readStart = readEnd
	}
	return nil
}

func IsNotExpired(runfolder string, expiredHours int) (bool, error) {
	notExpired := false
	err := filepath.Walk(runfolder, func(path string, info os.FileInfo, err error) error {
		if notExpired {
			return filepath.SkipDir
		}
		if err != nil {
			return filepath.SkipDir
		}

		if info == nil {
			return filepath.SkipDir
		}

		if time.Since(info.ModTime()).Hours() < float64(expiredHours) {
			notExpired = true
			fmt.Println(info.ModTime(), (time.Since(info.ModTime()).Hours() - float64(expiredHours)))
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		if err != filepath.SkipDir {

			return false, err
		}
	}

	return notExpired, nil
}

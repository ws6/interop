//parsing samplesheet
package samplesheetio

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var (
	//reasons for sqlite3
	ERR_EMPTY       = "EMPTY LINES"
	ERR_UNKNOWN     = "UNKNOWN SAMPLESHEET FMT"
	ERR_FMT_ISIS    = "NOT ISIS FMT"
	ERR_FMT_TAB     = "NOT TAB/CASAVA FMT"
	ERR_FMT_HEADER  = "FAILED PARSE HEADER"
	ERR_NO_SAMPLEID = "NO SAMPLE ID COLUM"
	ERR_LANE_NUM    = "LANE NUMBER ERROR"
	ERR_SAMPLE_NAME = "SAMPLE NAME EMPTY"
	ERR_UNALIGNED   = "HEADER AND ROW LENGTH MISMTACH"
	HEADER_DEF      = Header{
		SampleID:               HeaderDef{Accepts: []string{"SampleID", "Sample_ID", "Sample_Name"}},
		LaneNumber:             HeaderDef{Accepts: []string{"Lane", "Lanes"}},
		Index1:                 HeaderDef{Accepts: []string{"index"}},
		Index2:                 HeaderDef{Accepts: []string{"index2"}},
		I7_Index_ID:            HeaderDef{Accepts: []string{"I7_Index_ID"}},
		I5_Index_ID:            HeaderDef{Accepts: []string{"I5_Index_ID"}},
		Library_Prep_Lane:      HeaderDef{Accepts: []string{"Library_Prep_Lane", "LPLane"}},
		Library_Prep_Cartridge: HeaderDef{Accepts: []string{"LPCartridge"}},
		Manifest:               HeaderDef{Accepts: []string{"Manifest", "manifest"}},
	}
)

type LaneIndex struct {
	LaneNum                int
	SampleId               string
	IsIndexed              bool
	Barcode                []string // put them in order I7->I5
	BarcodeName            []string
	Library_Prep_Lane      string
	Library_Prep_Cartridge string
	Manifest               string
}

type HeaderDef struct {
	//Accepted Headers
	//SampleID,SampleName,index,index2,Lane,SampleProject,manifest,GenomeFolder,RunFolder
	//Lane,Sample_ID,Sample_Name,Sample_Plate,Sample_Well,I7_Index_ID,index,I5_Index_ID,index2,Sample_Project
	//
	Position      int      //1-offset to avoid default initialization
	Accepts       []string //any strings can accept
	StopWhenEmtpy bool
	Name          string //consolidated name
}

//ONLY support two index for now
type Header struct {
	SampleID               HeaderDef
	LaneNumber             HeaderDef
	Index1                 HeaderDef
	Index2                 HeaderDef
	I7_Index_ID            HeaderDef
	I5_Index_ID            HeaderDef
	Library_Prep_Lane      HeaderDef
	Library_Prep_Cartridge HeaderDef
	Manifest               HeaderDef
}
type SampleSheetLegcy struct {
	IsIsis     bool
	LaneIndex  []*LaneIndex
	Header     map[string]string
	Settings   map[string]string
	Manifests  map[string]string
	Content    []string
	Data       []map[string]string
	DataHeader []string
}

func detectPosition(hd HeaderDef, fields []string) int {
	for pos, f := range fields {
		for _, ac := range hd.Accepts {
			if strings.ToLower(ac) == strings.ToLower(f) {
				return pos
			}
		}
	}
	return -1
}

func ParseHeader(fields []string) Header {
	h := Header{}
	h.SampleID.Position = detectPosition(HEADER_DEF.SampleID, fields)
	h.LaneNumber.Position = detectPosition(HEADER_DEF.LaneNumber, fields)
	h.Index1.Position = detectPosition(HEADER_DEF.Index1, fields)
	h.Index2.Position = detectPosition(HEADER_DEF.Index2, fields)
	h.I7_Index_ID.Position = detectPosition(HEADER_DEF.I7_Index_ID, fields)
	h.I5_Index_ID.Position = detectPosition(HEADER_DEF.I5_Index_ID, fields)
	h.Library_Prep_Lane.Position = detectPosition(HEADER_DEF.Library_Prep_Lane, fields)
	h.Library_Prep_Cartridge.Position = detectPosition(HEADER_DEF.Library_Prep_Cartridge, fields)
	h.Manifest.Position = detectPosition(HEADER_DEF.Manifest, fields)
	return h
}
func ParseLine(header *Header, ln []string) (*LaneIndex, error) {
	ret := new(LaneIndex)

	ret.LaneNum = 1
	//NB if no header.LaneNumber.Position == -1, laneNumber = 1; this is for Miseq
	if header.LaneNumber.Position != -1 {
		num, err := strconv.Atoi(ln[header.LaneNumber.Position])
		if err != nil {
			return nil, err
		}
		ret.LaneNum = num
	}

	if header.SampleID.Position != -1 {
		ret.SampleId = ln[header.SampleID.Position]
		if strings.Trim(ret.SampleId, " ") == "" {
			return nil, fmt.Errorf(ERR_SAMPLE_NAME)
		}
	}
	//Library_Prep_Lane

	if header.Library_Prep_Lane.Position != -1 {
		ret.Library_Prep_Lane = ln[header.Library_Prep_Lane.Position]
		if strings.Trim(ret.Library_Prep_Lane, " ") == "" {
			return nil, fmt.Errorf(ERR_SAMPLE_NAME)
		}
	}
	//Library_Prep_Cartridge
	if header.Library_Prep_Cartridge.Position != -1 {
		ret.Library_Prep_Cartridge = ln[header.Library_Prep_Cartridge.Position]
		if strings.Trim(ret.Library_Prep_Cartridge, " ") == "" {
			return nil, fmt.Errorf(ERR_SAMPLE_NAME)
		}
	}
	//Manifest
	if header.Manifest.Position != -1 {
		ret.Manifest = ln[header.Manifest.Position]
		ret.Manifest = strings.Trim(ret.Manifest, " ") //okay if it is empty
	}

	if header.Index1.Position != -1 {
		ret.Barcode = append(ret.Barcode, ln[header.Index1.Position])
		barcodeName := "Index1"
		if header.I7_Index_ID.Position != -1 {
			barcodeName = ln[header.I7_Index_ID.Position]
		}
		ret.BarcodeName = append(ret.BarcodeName, barcodeName)
	}

	if header.Index2.Position != -1 {
		ret.Barcode = append(ret.Barcode, ln[header.Index2.Position])
		barcodeName := "Index2"
		if header.I5_Index_ID.Position != -1 {
			barcodeName = ln[header.I5_Index_ID.Position]
		}
		ret.BarcodeName = append(ret.BarcodeName, barcodeName)
	}

	if len(ret.Barcode) >= 0 {
		ret.IsIndexed = true
	}
	return ret, nil
}

//ParseIsisSection return a key-value map from header or settings sections
func ParseIsisSection(ss []string, sectionName string) (map[string]string, error) {
	dataStarted := false
	ret := make(map[string]string)
	for _, ln := range ss {
		if len(ln) == 0 {
			continue
		}
		if strings.HasPrefix(ln, sectionName) {
			dataStarted = true
			continue
		}

		if dataStarted && strings.HasPrefix(ln, "[") {

			break
		}

		if !dataStarted {
			continue
		}

		dataLn := strings.Split(ln, ",")
		lenData := len(dataLn)
		if lenData == 0 {
			continue
		}
		k := dataLn[0]
		if strings.Trim(k, " ") == "" {
			continue
		}
		v := ""
		if lenData >= 2 {
			v = dataLn[1]
		}
		ret[k] = v
	}

	return ret, nil
}

func ParseIsisSectionRow(ss []string, sectionName string) ([]string, error) {
	dataStarted := false
	ret := []string{}
	for _, ln := range ss {
		if len(ln) == 0 {
			continue
		}
		if strings.HasPrefix(ln, sectionName) {
			dataStarted = true
			continue
		}

		if dataStarted && strings.HasPrefix(ln, "[") {

			break
		}

		if !dataStarted {
			continue
		}

		ret = append(ret, ln)

	}

	return ret, nil
}

func (self *SampleSheetLegcy) ParseSectionRow(secName string) ([]string, error) {
	if self == nil {
		return nil, fmt.Errorf("SampleSheet is nil")
	}
	if self.Content == nil {
		return nil, fmt.Errorf("SampleSheet Content is nil")
	}

	return ParseIsisSectionRow(self.Content, secName)
}

func (self *SampleSheetLegcy) ParseSection(secName string) (map[string]string, error) {
	if self == nil {
		return nil, fmt.Errorf("SampleSheet is nil")
	}
	if self.Content == nil {
		return nil, fmt.Errorf("SampleSheet Content is nil")
	}

	return ParseIsisSection(self.Content, secName)
}

func ParseIsisSampleSheet(ss []string) (ret *SampleSheetLegcy, err error) {
	defer func() {
		if rerr := recover(); rerr != nil {
			err = fmt.Errorf("bad samplesheet\n%s", ss)
			ret = nil
		}
	}()
	ret = new(SampleSheetLegcy)
	ret.IsIsis = true
	ret.Content = ss
	dataStarted := false
	cnt := 0
	var header *Header
	headerLen := -1

	for _, ln := range ss {
		if len(ln) == 0 {
			continue
		}
		if strings.HasPrefix(ln, "[Data]") {
			dataStarted = true
			//parse data

			continue
		}
		if !dataStarted {
			continue
		}
		cnt++
		//Start processing
		dataLn := strings.Split(ln, ",")

		if cnt == 1 {
			hd := ParseHeader(dataLn)
			if hd.SampleID.Position == -1 {
				return nil, fmt.Errorf(ERR_NO_SAMPLEID)
			}
			header = &hd
			for _, v := range dataLn {
				if len(strings.Trim(v, " ")) != 0 {
					headerLen += 1
				}
			}

			ret.DataHeader = strings.Split(ln, ",")
			continue
			//parse header
		}

		if len(dataLn) >= len(ret.DataHeader) {
			m := make(map[string]string)
			for _i, k := range ret.DataHeader {
				m[k] = dataLn[_i]
			}
			ret.Data = append(ret.Data, m)
		}

		if len(dataLn) < headerLen {
			return nil, fmt.Errorf(ERR_UNALIGNED)
		}

		lane, err := ParseLine(header, dataLn)
		if err != nil {
			if err.Error() == ERR_SAMPLE_NAME {
				continue
			}
			return nil, fmt.Errorf("%s, %s", err.Error(), ln)
		}
		ret.LaneIndex = append(ret.LaneIndex, lane)
	}
	//	ret.Header, err = ParseIsisSection(ss, "[Header]")
	//	if err != nil {
	//		return nil, err
	//	}

	//	ret.Settings, err = ParseIsisSection(ss, "[Settings]")
	//	if err != nil {
	//		return nil, err
	//	}
	return ret, nil
}
func ParseTabSampleSheet(ss []string) (*SampleSheetLegcy, error) {
	ret := new(SampleSheetLegcy)
	ret.IsIsis = false
	ret.Content = ss
	cnt := 0
	var header *Header
	headerLen := -1
	for _, ln := range ss {

		cnt++
		//Start processing
		dataLn := strings.Split(ln, ",")
		_ = dataLn
		if cnt == 1 {
			hd := ParseHeader(dataLn)
			if hd.SampleID.Position == -1 {
				return nil, fmt.Errorf(ERR_NO_SAMPLEID)
			}
			header = &hd
			for _, v := range dataLn {
				if len(strings.Trim(v, " ")) != 0 {
					headerLen += 1
				}
			}
			continue
		}
		if len(dataLn) < headerLen {
			return nil, fmt.Errorf("ERR_UNALIGNED")

		}
		lane, err := ParseLine(header, dataLn)
		if err != nil {
			return nil, err

		}
		ret.LaneIndex = append(ret.LaneIndex, lane)
	}
	return ret, nil
}

func trimSamplSheetLine(s string) (bool, string) {
	r := strings.TrimRight(s, "\r ")
	//skip empty
	if len(r) == 0 {
		return false, r
	}
	//skip # comments
	if strings.HasPrefix(r, "#") {
		return false, r
	}
	r = strings.Trim(r, " ")
	//skip trimmed empty
	if len(r) == 0 {
		return false, r
	}
	return true, r
}
func ParseSampleSheet(sampleSheet string) (*SampleSheetLegcy, error) {
	//	ssb, err := ioutil.ReadFile(sampleSheetFile)
	//	if err != nil {
	//		return nil, err
	//	}
	ssSplit := strings.Split(string(sampleSheet), "\n")

	trimedSplit := []string{}
	for _, s := range ssSplit {
		if accepted, ts := trimSamplSheetLine(s); accepted {
			trimedSplit = append(trimedSplit, ts)
		}

	}

	if len(trimedSplit) == 0 {

		return nil, fmt.Errorf(ERR_EMPTY)
	}
	fmt.Println(trimedSplit[0])
	for _, b := range trimedSplit[0] {
		fmt.Println(b)
	}
	fmt.Println(strings.Contains(trimedSplit[0], `[Header]`))
	if strings.HasPrefix(trimedSplit[0], "[Header]") || strings.Contains(trimedSplit[0], `[Header]`) {
		fmt.Println(`trye Isis samplesheet`)
		return ParseIsisSampleSheet(trimedSplit)
	}

	if strings.HasPrefix(trimedSplit[0], "FCID") ||
		strings.HasPrefix(trimedSplit[0], "Lane") {
		return ParseTabSampleSheet(trimedSplit)
	}
	return nil, fmt.Errorf(ERR_UNKNOWN)
}

//GetSampleIndex this is not guranteed same as Isis.exe
func GetSampleIndex(ss *SampleSheetLegcy) []string {
	uniqMap := make(map[string]int)
	ret := []string{}
	for _, laneIdx := range ss.LaneIndex {
		if _, ok := uniqMap[laneIdx.SampleId]; !ok {
			uniqMap[laneIdx.SampleId] = 0
			ret = append(ret, laneIdx.SampleId)
		}
		uniqMap[laneIdx.SampleId]++
	}
	return ret
}

func ParseSS(expectedSampleSheet string) (*SampleSheetLegcy, error) {
	if _, err := os.Stat(expectedSampleSheet); os.IsNotExist(err) {
		return nil, fmt.Errorf("not SampleSheet.csv found %s", expectedSampleSheet)
	}
	//TODO check if samplesheet.csv is different from flowcell.samplesheet
	ssByte, err := ioutil.ReadFile(expectedSampleSheet)
	if err != nil {
		return nil, err
	}
	ss, err := ParseSampleSheet(string(ssByte))
	if err != nil {
		return nil, err
	}
	return ss, nil
}

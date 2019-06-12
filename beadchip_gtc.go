package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"sync"

	"io"
	"os"
)

var NO_ENTRY = fmt.Errorf(`NO ENTRY FOUND`)

//https://confluence.illumina.com/display/BIOINFO/GTC+File+Format

//beadchip_gtc parser out the .GTC file generated by  illumina beadarray tech
//If the ID in the TOC entry corresponds to an int variable type (NumSNPs or Ploidy, in our case), then the OFFSET in the TOC entry is the actual value of the unsigned integer and not a file offset. If the ID in the TOC entry corresponds to a string variable type, then the first byte at the file location specified by OFFSET is the length of the string (L) and the L bytes of the string follow as single byte characters. If the ID in the TOC corresponds to an array, then the first four bytes at the location specified by OFFSET are an integer corresponding to the length of the array (N).  Each element in the array follows. Arrays of type ushort have N ushorts.  Arrays of type byte have N bytes.  Arrays of type float have N floats.   The other two types require a bit of explanation.
const (
	ID_NUM_SNPS                 = 1
	ID_PLOIDY                   = 2
	ID_PLOIDY_TYPE              = 3
	ID_SAMPLE_NAME              = 10
	ID_SAMPLE_PLATE             = 11
	ID_SAMPLE_WELL              = 12
	ID_CLUSTER_FILE             = 100
	ID_SNP_MANIFEST             = 101
	ID_IMAGING_DATE             = 200
	ID_AUTOCALL_DATE            = 201
	ID_AUTOCALL_VERSION         = 300
	ID_NORMALIZATION_TRANSFORMS = 400
	ID_CONTROLS_X               = 500
	ID_CONTROLS_Y               = 501
	ID_RAW_X                    = 1000
	ID_RAW_Y                    = 1001
	ID_GENOTYPES                = 1002
	ID_BASE_CALLS               = 1003
	ID_GENOTYPE_SCORES          = 1004
	ID_SCANNER_DATA             = 1005
	ID_CALL_RATE                = 1006
	ID_GENDER                   = 1007
	ID_LOGR_DEV                 = 1008
	ID_GC10                     = 1009
	ID_GC50                     = 1011
	ID_B_ALLELE_FREQS           = 1012
	ID_LOGR_RATIOS              = 1013
	ID_PERCENTILES_X            = 1014
	ID_PERCENTILES_Y            = 1015
	ID_SLIDE_IDENTIFIER         = 1016 //Sentrix ID	string	The Sentrix barcode.
)

type Entry struct {
	ID     uint16
	Offset int32
}

type GTCHeader struct {
	//	Filename        string
	File            io.ReadSeeker
	Identifier      [3]byte
	Version         uint8
	NumberOfEntries int32
	Entries         []*Entry
}

func (self *GTCHeader) hasIdentity() bool {
	if self.Identifier[0] != 'g' {
		return false
	}
	if self.Identifier[1] != 't' {
		return false
	}
	if self.Identifier[2] != 'c' {
		return false
	}
	return true
}

func (self *GTCHeader) isVersionSupported() bool {
	versions := []uint8{3, 4, 5}
	for _, v := range versions {
		if self.Version == v {
			return true
		}
	}
	return false
}

func ParserGTCHeaderFromFile(filename string) (*GTCHeader, error) {
	file, err := os.Open(filename)
	if err != nil {

		return nil, err
	}
	defer file.Close()

	return ParserGTCHeader(file)
}

func ParserGTCHeader(file io.ReadSeeker) (*GTCHeader, error) {

	ret := new(GTCHeader)
	//	ret.Filename = filename
	ret.File = file
	buffer := bufio.NewReader(file)

	//read version
	if err := binary.Read(buffer, binary.LittleEndian, &ret.Identifier); err != nil {
		return nil, err
	}
	if !ret.hasIdentity() {
		return nil, fmt.Errorf(`not matching header identity %v`, ret.Identifier)
	}

	if err := binary.Read(buffer, binary.LittleEndian, &ret.Version); err != nil {
		return nil, err
	}
	if !ret.isVersionSupported() {
		return nil, fmt.Errorf(`version[%d] is not supported `, ret.Version)
	}

	if err := binary.Read(buffer, binary.LittleEndian, &ret.NumberOfEntries); err != nil {
		return nil, err
	}

	for i := 0; i < int(ret.NumberOfEntries); i++ {
		entry := Entry{}
		if err := binary.Read(buffer, binary.LittleEndian, &entry); err != nil {
			return nil, err
		}

		ret.Entries = append(ret.Entries, &entry)
	}
	return ret, nil
}

type Genotypes struct {
	Length int32
	Calls  []byte
}

type BaseCalls struct {
	Length int32
	Calls  []string
}

type Float32_T struct {
	Length int32
	Array  []float32
}

type BinCount struct {
	LeftCount  int //if less than left bound
	RightCount int //if greater than right bound
	Count      []int
	RawCount   []int
	Stdev      []float64 //deviation curve
	NumberBins int
	Left       float32
	Right      float32
	Bin        float32
}

func (self *Float32_T) Bin(left, right, bin float32) (*BinCount, error) {
	if left > right {
		left, right = right, left
	}
	ret := new(BinCount)
	ret.Left = left
	ret.Right = right
	ret.Bin = bin
	ret.NumberBins = int((right - left) / bin)
	if ret.NumberBins <= 0 {
		return nil, fmt.Errorf(`bin size too large`)
	}

	ret.Count = make([]int, ret.NumberBins)

	for _, f := range self.Array {
		if f < left {
			ret.LeftCount++
			continue
		}
		if f > right {
			ret.RightCount++

			continue
		}
		idx := int((f - left) / bin)
		if idx >= ret.NumberBins {
			idx = ret.NumberBins - 1
		}
		if idx < 0 {
			idx = 0
		}
		ret.Count[idx]++
	}

	return ret, nil
}

func (self *GTCHeader) ParseGenotypes() (*Genotypes, error) {

	return self.parseGenotypes(self.File)
}

func (self *GTCHeader) parseFloat32_T(file io.ReadSeeker, id int) (*Float32_T, error) {
	ret := new(Float32_T)
	offset := int64(0)
	for _, entry := range self.Entries {
		if entry.ID == uint16(id) {
			offset = int64(entry.Offset)
		}
	}
	if offset == 0 {
		return nil, NO_ENTRY
	}

	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}
	buffer := bufio.NewReader(file)

	if err := binary.Read(buffer, binary.LittleEndian, &ret.Length); err != nil {
		return nil, err
	}

	for i := 0; i < int(ret.Length); i++ {
		var b float32
		if err := binary.Read(buffer, binary.LittleEndian, &b); err != nil {
			return nil, err
		}

		ret.Array = append(ret.Array, b)
	}
	return ret, nil
}

func (self *GTCHeader) ParseBalleleFreq() (*Float32_T, error) {

	return self.parseFloat32_T(self.File, ID_B_ALLELE_FREQS)
}
func (self *GTCHeader) ParseLogRRatio() (*Float32_T, error) {

	return self.parseFloat32_T(self.File, ID_LOGR_RATIOS)
}

func (self *GTCHeader) ParseGenotypeScores() (*Float32_T, error) {

	return self.parseFloat32_T(self.File, ID_GENOTYPE_SCORES)
}

func (self *GTCHeader) parseGenotypes(file io.ReadSeeker) (*Genotypes, error) {
	ret := new(Genotypes)

	offset, err := self.getOffset(ID_GENOTYPES)
	if err != nil {
		return nil, err
	}
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}
	buffer := bufio.NewReader(file)

	if err := binary.Read(buffer, binary.LittleEndian, &ret.Length); err != nil {
		return nil, err
	}

	for i := 0; i < int(ret.Length); i++ {
		var b byte
		if err := binary.Read(buffer, binary.LittleEndian, &b); err != nil {
			return nil, err
		}

		ret.Calls = append(ret.Calls, b)
	}
	return ret, nil
}

func ParseGTCGenotypes(filename string) (*Genotypes, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	gtcHeader, err := ParserGTCHeader(file)
	if err != nil {
		return nil, err
	}

	file.Seek(0, 0)
	return gtcHeader.parseGenotypes(file)

}

func score(s1, s2 *Genotypes) float64 {
	size := len(s1.Calls)
	if size != len(s2.Calls) {
		return 0
	}
	matched := 0
	for i, _ := range s1.Calls {
		if s1.Calls[i] == s2.Calls[i] {
			matched++
		}
	}
	return float64(float64(100*matched) / float64(size))

}

func Score(gtc1, gtc2 string) (float64, error) {
	var wg sync.WaitGroup
	wg.Add(2)
	var s1, s2 *Genotypes
	var err error
	go func() {

		defer wg.Done()
		_s1, _err := ParseGTCGenotypes(gtc1)
		if _err != nil {
			err = _err
			return
		}
		s1 = _s1
	}()
	go func() {

		defer wg.Done()
		_s2, _err := ParseGTCGenotypes(gtc2)
		if _err != nil {
			err = _err
			return
		}
		s2 = _s2
	}()
	if err != nil {
		return 0, err
	}
	wg.Wait()

	return score(s1, s2), nil
}

type GTCInfo struct {
	SampleName                  string
	SamplePlate                 string
	SampleWell                  string
	ClusterFile                 string
	SNPManifest                 string
	ImagingDate                 string
	AutoCallDate                string
	AutoCallVersion             string
	PloidyType                  string
	ploidyType, Ploidy, NumSNPs int32
	//	1	Diploid
	//2	Autopolyploid
	//3	Allopolyploid

	SentrixID string //The Sentrix barcode.
	LogRDev   float32

	EstimatedGender string

	ScannerName string
	CallRate    float32

	//	Scanner     *ScannerInfo
}

func (self *GTCHeader) readInt32(id uint16) (int32, error) {

	for _, entry := range self.Entries {
		if entry.ID == id {
			return entry.Offset, nil

		}
	}
	return 0, NO_ENTRY
}

func (self *GTCHeader) readString(file io.ReadSeeker, id uint16) (string, error) {

	offset, err := self.getOffset(id)
	if err != nil {
		return "", err
	}
	return self.readStringByOffset(file, offset)

}

func (self *GTCHeader) getOffset(id uint16) (int64, error) {
	offset := int64(0)
	for _, entry := range self.Entries {
		if entry.ID == id {
			offset = int64(entry.Offset)
		}
	}
	if offset == 0 {
		return 0, NO_ENTRY
	}

	return offset, nil
}

func (self *GTCHeader) readFloat32(file io.ReadSeeker, id uint16) (float32, error) {

	offset, err := self.getOffset(id)
	if err != nil {
		return 0., err
	}
	if _, err := file.Seek(offset, 0); err != nil {
		return 0., err
	}

	buffer := bufio.NewReader(file)
	var ret float32
	if err := binary.Read(buffer, binary.LittleEndian, &ret); err != nil {
		return 0., err
	}

	return ret, nil

}

func (self *GTCHeader) readChar(file io.ReadSeeker, id uint16) (string, error) {

	offset, err := self.getOffset(id)
	if err != nil {
		return "", err
	}
	if _, err := file.Seek(offset, 0); err != nil {
		return "", err
	}

	buffer := bufio.NewReader(file)
	var ret byte
	if err := binary.Read(buffer, binary.LittleEndian, &ret); err != nil {
		return "", err
	}

	return string(ret), nil

}

func (self *GTCHeader) readStringByOffset(file io.ReadSeeker, offset int64) (string, error) {
	if _, err := file.Seek(offset, 0); err != nil {
		return "", err
	}

	buffer := bufio.NewReader(file)
	var size uint8
	if err := binary.Read(buffer, binary.LittleEndian, &size); err != nil {
		return "", err
	}

	ret := []byte{}

	for i := 0; i < int(size); i++ {
		var b byte
		if err := binary.Read(buffer, binary.LittleEndian, &b); err != nil {
			return "", err
		}

		ret = append(ret, b)
	}
	return string(ret), nil
}

type ScannerInfo struct {
	ScannerName    string
	PmtGreen       int32
	PmtRed         int32
	ScannerVersion string
	ImagingUser    string
}

func (self *GTCHeader) GetScannerInfo(file *os.File) (*ScannerInfo, error) {
	ret := new(ScannerInfo)
	offset, err := self.getOffset(ID_SCANNER_DATA)
	if err != nil {
		return nil, err
	}
	ret.ScannerName, _ = self.readStringByOffset(file, offset)

	var pmtGreen int32
	//!!offset moved
	buffer := bufio.NewReader(file)
	if err := binary.Read(buffer, binary.LittleEndian, &pmtGreen); err != nil {
		return nil, err
	}

	ret.PmtGreen = pmtGreen

	var pmtRed int32
	if err := binary.Read(buffer, binary.LittleEndian, &pmtRed); err != nil {
		return nil, err
	}

	ret.PmtRed = pmtRed
	_offset := int64(0)
	_offset = offset + int64(len(ret.ScannerName)) + int64(8)

	ret.ScannerVersion, _ = self.readStringByOffset(file, _offset)
	_offset += int64(len(ret.ScannerVersion))
	ret.ImagingUser, _ = self.readStringByOffset(file, _offset)
	return ret, nil
}

func (self *GTCHeader) GetGTCInfo() (*GTCInfo, error) {
	file := self.File
	ret := new(GTCInfo)
	//	ret.Scanner, err = self.GetScannerInfo(file)
	//	if err != nil {
	//		return nil, err
	//	}
	var err error
	ret.SampleName, _ = self.readString(file, ID_SAMPLE_NAME)
	ret.SamplePlate, _ = self.readString(file, ID_SAMPLE_PLATE)
	ret.SampleWell, _ = self.readString(file, ID_SAMPLE_WELL)
	ret.ClusterFile, _ = self.readString(file, ID_CLUSTER_FILE)
	ret.SNPManifest, _ = self.readString(file, ID_SNP_MANIFEST)
	ret.ImagingDate, _ = self.readString(file, ID_IMAGING_DATE)
	ret.AutoCallDate, _ = self.readString(file, ID_AUTOCALL_DATE)
	ret.AutoCallVersion, _ = self.readString(file, ID_AUTOCALL_VERSION)
	ret.SentrixID, _ = self.readString(file, ID_SLIDE_IDENTIFIER)
	ret.ScannerName, _ = self.readString(file, ID_SCANNER_DATA)

	//	PloidyType, Ploidy, NumSNPs int32
	ret.NumSNPs, _ = self.readInt32(ID_NUM_SNPS)
	ret.ploidyType, err = self.readInt32(ID_PLOIDY_TYPE)
	if err == nil {
		if ret.ploidyType == 1 {
			ret.PloidyType = `Diploid`
		}
		if ret.ploidyType == 2 {
			ret.PloidyType = `Autopolyploid`
		}
		if ret.ploidyType == 3 {
			ret.PloidyType = `Allopolyploid`
		}
	}
	ret.Ploidy, _ = self.readInt32(ID_PLOIDY)
	ret.CallRate, _ = self.readFloat32(file, ID_CALL_RATE)
	ret.LogRDev, _ = self.readFloat32(file, ID_LOGR_DEV)
	ret.EstimatedGender, _ = self.readChar(file, ID_GENDER)
	return ret, nil
}

func (self *GTCHeader) ParseBaseCalls() (*BaseCalls, error) {

	return self.parseBaseCalls(self.File)
}

func (self *GTCHeader) parseBaseCalls(file io.ReadSeeker) (*BaseCalls, error) {
	gtcHeader, err := self.GetGTCInfo()
	if err != nil {
		return nil, err
	}
	if gtcHeader.ploidyType != 1 {
		return nil, fmt.Errorf(`not Diploid is not implemented`)
	}
	ret := new(BaseCalls)

	offset, err := self.getOffset(ID_BASE_CALLS)
	if err != nil {
		return nil, err
	}
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}
	buffer := bufio.NewReader(file)

	if err := binary.Read(buffer, binary.LittleEndian, &ret.Length); err != nil {
		return nil, err
	}

	for i := 0; i < int(ret.Length); i++ {
		var b [2]byte
		if err := binary.Read(buffer, binary.LittleEndian, &b); err != nil {
			return nil, err
		}

		ret.Calls = append(ret.Calls, string(b[0])+string(b[1]))
	}
	return ret, nil
}

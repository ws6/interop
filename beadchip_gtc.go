package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"sync"

	"os"
)

//beadchip_gtc parser out the .GTC file generated by  illumina beadarray tech

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
	ID_SLIDE_IDENTIFIER         = 1016
)

type Entry struct {
	ID     uint16
	Offset int32
}

type GTCHeader struct {
	Filename        string
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

func ParserGTCHeader(filename string) (*GTCHeader, error) {
	file, err := os.Open(filename)
	if err != nil {

		return nil, err
	}
	defer file.Close()
	return parserGTCHeader(filename, file)
}

func parserGTCHeader(filename string, file *os.File) (*GTCHeader, error) {
	ret := new(GTCHeader)
	ret.Filename = filename

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

func (self *GTCHeader) ParseGenotypes() (*Genotypes, error) {
	file, err := os.Open(self.Filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return self.parseGenotypes(file)
}

func (self *GTCHeader) parseGenotypes(file *os.File) (*Genotypes, error) {
	ret := new(Genotypes)
	offset := int64(0)
	for _, entry := range self.Entries {
		if entry.ID == ID_GENOTYPES {
			offset = int64(entry.Offset)
		}
	}
	if offset == 0 {
		return nil, fmt.Errorf(`no ID_GENOTYPES found from entries`)
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
	gtcHeader, err := ParserGTCHeader(filename)
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

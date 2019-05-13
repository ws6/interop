package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const HEADER_BPM = `BPM`
const (
	REFSTRAND_UNKNOWN = 0
	REFSTRAND_PLUS    = 1
	REFSTRAND_MINUS   = 2

	SOURCESTRAND_UNKNOWN              = 0
	SOURCESTRAND_SOURCESTRAND_FORWARD = 1
	SOURCESTRAND_REVERSE              = 2
)

type BPMInfo struct {
	Version          int32
	NumerOfLoci      int32
	ManifestName     string
	ControlString    string
	LociNames        []string
	NormalizationIds []uint8
	LocusEntries     []*LocusEntry
	LocusEntryMap    map[string]*LocusEntry
}

type BPMParseOptions struct {
	OnlyHeader          bool
	OnlyLociName        bool
	OnlyNormalizationId bool
}

type LocusEntry struct {
	Version   uint32
	IlmnId    string
	Name      string
	SNP       string
	Chrom     string
	MapInfo   int
	AssayType int
	AddressA  int32
	AddressB  int32

	RefStrand          int
	RefStrandString    string
	SourceStrand       int
	SourceStrandString string
}

func NewLocusEntry() *LocusEntry {
	ret := new(LocusEntry)
	ret.RefStrand = REFSTRAND_UNKNOWN
	ret.SourceStrand = SOURCESTRAND_UNKNOWN
	ret.MapInfo = -1
	ret.AssayType = -1
	ret.AddressA = -1
	ret.AddressB = -1
	return ret
}

//readMSString  Helper function to parse string from file handle. See https://msdn.microsoft.com/en-us/library/yzxa6408(v=vs.100).aspx
func readMSString(buffer *bufio.Reader) (string, error) {
	totalLength := 0
	numBytes := uint(0)
	_partial, err := buffer.ReadByte()
	if err != nil {
		return "", err
	}
	for {
		if _partial&0x80 == 0 {
			break
		}
		totalLength += int(_partial&0x7F) << (7 * numBytes)
		_partial, err = buffer.ReadByte()
		if err != nil {
			return "", err
		}
		numBytes++
	}
	totalLength += int(_partial) << (7 * numBytes)
	bret := make([]byte, totalLength)
	for i := 0; i < totalLength; i++ {

		b, err := buffer.ReadByte()
		if err != nil {
			return "", err
		}

		bret[i] = b
	}

	return string(bret), nil
}

func readLocusEntry6(ret *LocusEntry, buffer *bufio.Reader) error {
	var err error
	ret.IlmnId, err = readMSString(buffer)
	if err != nil {
		return err
	}

	ilmnId := strings.Split(ret.IlmnId, "_")
	if len(ilmnId) >= 2 {
		ret.SourceStrandString = ilmnId[len(ilmnId)-2]
	}

	ret.Name, err = readMSString(buffer)
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		if _, err = readMSString(buffer); err != nil {
			return err
		}
	}
	for i := 0; i < 4; i++ {
		if _, err := buffer.ReadByte(); err != nil {
			return err
		}
	}
	for i := 0; i < 2; i++ {
		if _, err = readMSString(buffer); err != nil {
			return err
		}
	}
	ret.SNP, err = readMSString(buffer)

	if err != nil {
		return err
	}
	ret.Chrom, err = readMSString(buffer)

	if err != nil {
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = readMSString(buffer); err != nil {
			return err
		}
	}
	mapInfoStr, err := readMSString(buffer)

	ret.MapInfo, err = strconv.Atoi(mapInfoStr)
	if err != nil {
		return fmt.Errorf(`mapInfoStr Atoi err:%s`, err.Error())
	}

	for i := 0; i < 2; i++ {
		if _, err = readMSString(buffer); err != nil {
			return err
		}
	}
	if err := binary.Read(buffer, binary.LittleEndian, &ret.AddressA); err != nil {
		return err
	}
	if err := binary.Read(buffer, binary.LittleEndian, &ret.AddressB); err != nil {
		return err
	}

	for i := 0; i < 7; i++ {
		if _, err = readMSString(buffer); err != nil {
			return err
		}
	}
	for i := 0; i < 3; i++ {
		if _, err := buffer.ReadByte(); err != nil {
			return err
		}
	}
	bAssayType, err := buffer.ReadByte()
	if err != nil {
		return err
	}
	ret.AssayType = int(bAssayType)

	return nil

}

func readLocusEntry7(ret *LocusEntry, buffer *bufio.Reader) error {
	if err := readLocusEntry6(ret, buffer); err != nil {
		return err
	}

	for i := 0; i < 16; i++ {
		if _, err := buffer.ReadByte(); err != nil {
			return err
		}
	}
	return nil
}

func readLocusEntry8(ret *LocusEntry, buffer *bufio.Reader) error {

	if err := readLocusEntry7(ret, buffer); err != nil {
		return err
	}
	var err error
	ret.RefStrandString, err = readMSString(buffer)
	if err != nil {
		return err
	}
	return nil

}

func readLocusEntry(buffer *bufio.Reader) (*LocusEntry, error) {
	ret := NewLocusEntry()
	if err := binary.Read(buffer, binary.LittleEndian, &ret.Version); err != nil {
		return nil, err
	}

	if ret.Version < 6 || ret.Version > 8 {
		return nil, fmt.Errorf(`  "Manifest format error: unknown version for locus entry [%d]`, ret.Version)
	}
	if ret.Version == 6 {
		if err := readLocusEntry6(ret, buffer); err != nil {
			return nil, err
		}

	}
	if ret.Version == 7 {
		if err := readLocusEntry7(ret, buffer); err != nil {
			return nil, err
		}

	}
	if ret.Version == 8 {
		if err := readLocusEntry8(ret, buffer); err != nil {
			return nil, err
		}

	}
	return ret, nil

}

func ParseBMP(file io.ReadSeeker, opts *BPMParseOptions) (*BPMInfo, error) {
	ret := new(BPMInfo)
	buffer := bufio.NewReader(file)

	//	//read version
	//	if err := binary.Read(buffer, binary.LittleEndian, &ret.Identifier); err != nil {
	//		return nil, err
	//	}
	header := ""
	for i := 0; i < 3; i++ {
		r, _, err := buffer.ReadRune()
		if err != nil {
			return nil, err
		}
		header += string(r)

	}
	if header != HEADER_BPM {
		return nil, fmt.Errorf(`wrong header, not BPM`)
	}
	bversion, err := buffer.ReadByte()
	if err != nil {
		return nil, err
	}
	if bversion != 1 {
		return nil, fmt.Errorf(`bversion != 1 `)
	}

	if err := binary.Read(buffer, binary.LittleEndian, &ret.Version); err != nil {
		return nil, err
	}

	version_flag := int32(0x1000)
	if ret.Version&version_flag == version_flag {
		ret.Version = ret.Version ^ version_flag
	}

	if ret.Version > 5 || ret.Version < 3 {
		return nil, fmt.Errorf(`unsupported BMP version %d`, ret.Version)
	}
	//read string
	ret.ManifestName, err = readMSString(buffer)
	if err != nil {
		return nil, fmt.Errorf(`readMSString ManifestName err:%s`, err.Error())
	}
	if ret.Version > 1 {
		ret.ControlString, err = readMSString(buffer)
		if err != nil {
			return nil, err
		}
	}

	if err := binary.Read(buffer, binary.LittleEndian, &ret.NumerOfLoci); err != nil {
		return nil, err
	}
	//
	if opts.OnlyHeader {
		return ret, nil
	}

	//	file.Seek(int64(4*ret.NumerOfLoci), io.SeekCurrent)
	for i := 0; i < 4*int(ret.NumerOfLoci); i++ {
		buffer.ReadByte()
		//		buffer.ReadByte()
		//		buffer.ReadByte()
		//		buffer.ReadByte()
	}
	for i := 0; i < int(ret.NumerOfLoci); i++ {
		loci, err := readMSString(buffer)
		if err != nil {
			return nil, err
		}
		ret.LociNames = append(ret.LociNames, loci)

	}
	if opts.OnlyLociName {
		return ret, nil
	}
	ret.NormalizationIds = make([]uint8, int(ret.NumerOfLoci))
	for i := 0; i < int(ret.NumerOfLoci); i++ {
		normId, err := buffer.ReadByte()
		if err != nil {
			return nil, err
		}
		if uint8(normId) >= 100 {
			return nil, fmt.Errorf(`Manifest format error: read invalid normalization ID`)
		}
		ret.NormalizationIds[i] = uint8(normId)
	}

	if opts.OnlyNormalizationId {
		return ret, nil
	}

	for i := 0; i < int(ret.NumerOfLoci); i++ {
		//readLocusEntry
		entry, err := readLocusEntry(buffer)
		if err != nil {
			return nil, err
		}

		ret.LocusEntries = append(ret.LocusEntries, entry)
	}
	ret.LocusEntryMap = make(map[string]*LocusEntry)
	for _, le := range ret.LocusEntries {
		ret.LocusEntryMap[le.Name] = le
	}
	//!!!sanity check
	for _, name := range ret.LociNames {
		if _, ok := ret.LocusEntryMap[name]; !ok {
			return nil, fmt.Errorf(`Assay [%s] doesn't have entry info`, name)
		}
	}
	return ret, nil
}

func ParseBMPNames(file io.ReadSeeker) (*BPMInfo, error) {
	opts := new(BPMParseOptions)
	opts.OnlyLociName = true
	return ParseBMP(file, opts)
}

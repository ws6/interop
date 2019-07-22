package interop

import (
	"fmt"
	"io"
	"os"

	"strings"
)

func (self *GTCHeader) ToPED(_sampleName string) ([]string, error) {
	gtcInfo, err := self.GetGTCInfo()
	if err != nil {
		return nil, err
	}
	sex := "0"

	switch gtcInfo.EstimatedGender {
	case "M":
		sex = "1"
	case "F":
		sex = "2"
	default:
		sex = "0"

	}
	sampleName := gtcInfo.SampleName
	if _sampleName != "" {
		sampleName = _sampleName
	}
	sampleName = strings.Replace(sampleName, " ", "_", -1)
	ret := []string{
		"0", //family_id
		sampleName,
		"0", //paternal_id
		"0", //maternal_id
		sex,
		"0", //affection
	}

	basecalls, err := self.ParseBaseCalls()
	if err != nil {
		return nil, err
	}
	for i, s := range basecalls.Calls {
		if len(s) != 2 {
			return nil, fmt.Errorf(`base call is not size of 2 at pos[%d] got->[%s]`, i, s)
		}
		_s := strings.Replace(s, "-", "0", -1)
		ret = append(ret, string(_s[0]))
		ret = append(ret, string(_s[1]))
	}
	//	ret = append(ret, "\n") //!!! different from os.linesep from python lib
	return ret, nil
}

type PEDBaseCall struct {
	Allele1 string
	Allele2 string
}

type PED struct {
	FamilyId   string
	SampleName string
	PaternalId string
	MaternalId string
	Sex        string
	Affection  string
	BaseCalls  []*PEDBaseCall
	NumLoci    int
}

func ParsePED(pedStr string) (*PED, error) {
	elements := strings.Split(pedStr, " ")
	if len(elements) < 6 {
		return nil, fmt.Errorf(`not enough header info <6`)
	}
	sz := len(elements)
	if len(elements[sz-1]) > 1 { //trim newline
		elements[sz-1] = string(elements[sz-1][0])
	}
	ret := new(PED)
	ret.FamilyId = elements[0]
	ret.SampleName = elements[1]
	ret.PaternalId = elements[2]
	ret.MaternalId = elements[3]
	ret.Sex = elements[4]
	ret.Affection = elements[5]
	if (sz-6)%2 == 1 {
		return nil, fmt.Errorf(`odd number alleles[%d]`, (sz - 6))
	}

	numLoci := (sz - 6) / 2
	if numLoci <= 0 {
		return ret, nil
	}

	ret.NumLoci = numLoci

	ret.BaseCalls = make([]*PEDBaseCall, numLoci)
	for i := 0; i < numLoci; i++ {
		toPush := new(PEDBaseCall)
		toPush.Allele1 = string(elements[6:][i*2])
		toPush.Allele2 = string(elements[6:][i*2+1])
		ret.BaseCalls[i] = toPush
	}
	return ret, nil
}

func ParsePEDBaseCall(pedStr string) (*BaseCalls, error) {
	ret := new(BaseCalls)
	return ret, nil
}

func (self *PED) Callrate() float64 {
	sz := len(self.BaseCalls)
	if sz == 0 {
		return 0.
	}
	numNoCall := 0
	for _, bc := range self.BaseCalls {
		if bc.Allele1 == "0" || bc.Allele2 == "0" {
			numNoCall++
			continue
		}
		if bc.Allele1 == "-" || bc.Allele2 == "-" {
			numNoCall++
			continue
		}
	}

	return 1. - float64(numNoCall)/float64(sz)

}

func validCharset(s string) bool {
	switch s {
	case "A":
		return true
	case "G":
		return true
	case "C":
		return true
	case "T":
		return true
	case "I":
		return true
	case "D":
		return true
	case "0":
		return true
	}

	return false

}

func (self *PED) ValidCharset() bool {
	for _, bc := range self.BaseCalls {
		if !validCharset(bc.Allele1) {
			return false
		}
		if !validCharset(bc.Allele2) {
			return false
		}
	}

	return true
}

func (self *PED) Print(f io.Writer) error {
	ret := []string{
		self.FamilyId,
		self.SampleName,
		self.PaternalId,
		self.MaternalId,
		self.Sex,
		self.Affection,
	}
	_, err := f.Write([]byte(strings.Join(ret, " ")))
	if err != nil {
		return fmt.Errorf(`writting header error:%s`, err.Error())
	}
	for _, s := range self.BaseCalls {
		if _, err := f.Write([]byte(" " + s.Allele1)); err != nil {
			return err
		}
		if _, err := f.Write([]byte(" " + s.Allele2)); err != nil {
			return err
		}
	}
	//put a newline at the end for consistency
	if _, err := f.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func (self *PED) PrintToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return self.Print(f)
}

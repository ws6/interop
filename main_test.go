package interop

import (
	"testing"
)

func TestGTCParser(t *testing.T) {
	filename := `\\ussd-prd-isi05\fts_genotyping\call_file\20180418\202428760007_R12C02.gtc`
	h, err := ParserGTCHeader(filename)
	if err != nil {
		t.Fatal(err.Error())
	}

	gt, err := h.ParseGenotypes()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("genotypes %+v", gt.Length)
}

//ParseGTCGenotypes
func TestParseGenotyps(t *testing.T) {
	filename := `\\ussd-prd-isi05\fts_genotyping\call_file\20180418\202428760007_R12C02.gtc`
	gt, err := ParseGTCGenotypes(filename)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("genotypes %+v", len(gt.Calls))
}

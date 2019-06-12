package interop

import (
	"testing"
)

func _TestGTCScore(t *testing.T) {
	gtc1 := `\\ussd-prd-isi05\fts_genotyping\call_file\20180321\202351440138_R06C01.gtc`
	gtc2 := `\\ussd-prd-isi05\fts_genotyping\call_file\20180419\202238540120_R03C01.gtc`
	m, err := Score(gtc1, gtc2)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf(`matched score %f`, m)
}

//ParseGTCGenotypes
func _TestParseGenotyps(t *testing.T) {
	filename := `\\ussd-prd-isi05\fts_genotyping\call_file\20180418\202428760007_R12C02.gtc`
	gt, err := ParseGTCGenotypes(filename)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("genotypes %+v", len(gt.Calls))
}

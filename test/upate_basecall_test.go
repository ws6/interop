package test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ws6/interop"
)

func xTestGTCToPed(t *testing.T) {
	filename := `\\ussd-prd-isi05\fts_laboratory\Health_Screening_Proficiency_Testing_ILS-SD\203564140035_R12C02.gtc`
	file, err := os.Open(filename)
	if err != nil {

		t.Fatal(err.Error())
	}
	defer file.Close()

	gtcHeader, err := interop.ParserGTCHeader(file)
	if err != nil {
		t.Fatal(err.Error())
	}
	pedStr, err := gtcHeader.ToPED("")
	if err != nil {
		t.Fatal(err.Error())
	}
	ped, err := interop.ParsePED(string(strings.Join(pedStr, " ")))
	if err != nil {
		t.Fatal(err.Error())
	}

	ped.PrintToFile(`203564140035_R12C02.ped`)

}

func xTestUpdatePEDBasecall(t *testing.T) {
	pedFileName := `473ea65d-5ae1-4923-8401-b51675ec21f8.ped`
	pedBytes, err := ioutil.ReadFile(pedFileName)
	if err != nil {
		t.Fatal(err.Error())
	}
	ped, err := interop.ParsePED(string(pedBytes))
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf(`callrate=%f  [%s]`, ped.Callrate(), pedFileName)
	return
	var toUpdates = []struct {
		pos     int
		allele1 string
		allele2 string
	}{

		{1593, "D", "D"},
		{1594, "D", "D"},
		{1595, "D", "D"},
		{1596, "D", "D"},
		{309903, "D", "D"},
		{309904, "D", "D"},
		{2967, "A", "A"},
		{124083, "A", "A"},
		{9153, "A", "A"},
		{9154, "A", "A"},
		{218412, "A", "A"},
		{2962, "G", "G"},
		{2963, "G", "G"},
		{2964, "G", "G"},
		{2965, "G", "G"},
		{3016, "A", "A"},
		{3017, "A", "A"},
		{3018, "A", "A"},
		{124168, "A", "A"},
	}

	for _, u := range toUpdates {
		ped.BaseCalls[u.pos].Allele1 = u.allele1
		ped.BaseCalls[u.pos].Allele2 = u.allele2

	}
	//output
	outputPedName := `All_PP_five_v10.ped`
	if err := ped.PrintToFile(outputPedName); err != nil {
		t.Fatal(err.Error())
	}
	//shortcut to load ped from file
	ped2, err := interop.ParsePEDFromFile(outputPedName)
	if err != nil {
		t.Fatal(err.Error())
	}
	ped1, err := interop.ParsePEDFromFile(pedFileName)
	if err != nil {
		t.Fatal(err.Error())
	}
	if ped2.SameBaseCalls(ped1) {
		t.Error(`should not same ped basecalls`)
		t.Fail()
	}

}

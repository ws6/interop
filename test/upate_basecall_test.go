package test

import (
	"io/ioutil"

	"testing"

	"github.com/ws6/interop"
)

func TestUpdatePEDBasecall(t *testing.T) {
	pedFileName := `203435540157_R01C01.ped`
	pedBytes, err := ioutil.ReadFile(pedFileName)
	if err != nil {
		t.Fatal(err.Error())
	}
	ped, err := interop.ParsePED(string(pedBytes))
	if err != nil {
		t.Fatal(err.Error())
	}
	var toUpdates = []struct {
		pos     int
		allele1 string
		allele2 string
	}{
		{101, "A", "A"}, //change to whatever you want
		{2001, "T", "A"},
	}

	for _, u := range toUpdates {
		ped.BaseCalls[u.pos].Allele1 = u.allele1
		ped.BaseCalls[u.pos].Allele2 = u.allele2

	}
	//output
	outputPedName := `after_update.ped`
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

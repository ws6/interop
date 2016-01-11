package interop

import (
	"testing"
)

func TestSubTileStat(t *testing.T) {
	//	pfGridFilename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\160106_HWI-ST347_1760_BHKFGGCCXX\InterOp\PFGridMetricsOut.bin`
	pfGridFilename := `\\ussd-prd-isi04\Voyager\160108_HS3000-1132_2191_Bh2gthcfxx\InterOp\PFGridMetricsOut.bin`
	//	fwhmFilename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\160106_HWI-ST347_1760_BHKFGGCCXX\InterOp\FWHMGridMetricsOut.bin`
	fwhmFilename := `\\ussd-prd-isi04\Voyager\160108_HS3000-1132_2191_Bh2gthcfxx\InterOp\FWHMGridMetricsOut.bin`

	subtileInfo := new(SubtileInfo)
	pf := PFMetricsInfo{Filename: pfGridFilename}
	if err := pf.Parse(); err != nil {
		t.Error(err)
	}
	subtileInfo.PFInfo = &pf
	fwhm := FwhmMetricsInfo{Filename: fwhmFilename}
	if err := fwhm.Parse(); err != nil {
		t.Error(err)
	}

	subtileInfo.FwhmInfo = &fwhm

	if err := subtileInfo.MakeBoxStat(); err != nil {
		t.Fatal(err.Error())
	}
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.ClusterPF)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.ClusterRaw)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.DensityPF)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.DensityRaw)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.DensityPF)
	for k, v := range subtileInfo.SubtileLaneStat.PF.XBinStat {
		t.Logf("%d  \n", k)
		for bin, stat := range v {
			t.Logf("%d %+v\n", bin, stat)
		}
	}
	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.PF.XBinStat)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.PF.YBinStat)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.FWHMAll_Channel_All)
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.FWHM_Channels[0])
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.FWHM_Channels[1])
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.FWHM_Channels[2])
	//	t.Logf("%+v\n", subtileInfo.SubtileLaneStat.FWHM_Channels[3])

}

func TestBoxStat(t *testing.T) {
	a := []float64{1, 2, 3, 4, 5, 6, 7, 12, 3, 5, 19, 7}
	stat := new(BoxWhiskerStat)
	if err := stat.GetFloat64(&a); err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%+v\n", stat)

}

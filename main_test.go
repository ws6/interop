package interop

import (
	"encoding/json"

	"testing"
)

func TestEmpericalPhasing(t *testing.T) {
	filename := `\\ussd-prd-isi04\Voyager\160701_VP1-08_0164_A027BCABVY\InterOp\EmpiricalPhasingMetricsOut.bin`

	em := EmpericalPhasingInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(em.Version, em.SSize, len(em.Metrics))
	printed := 0
	for _, _m := range em.Metrics {
		if printed > 20 {
			break
		}
		if _m.Phasing > 0 {
			t.Logf("%+v\n", _m)
			printed++
		}

	}

}
func TestExtendMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/TileMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\TileMetricsOut.bin`
	filename := `\\ussd-prd-isi04\Voyager\160401_SN924_2689_A057KBAAVY\InterOp\ExtendedTileMetricsOut.bin`
	em := ExtendMetricsInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log(em.Version, em.SSize)
	for _, v := range em.Metrics {
		//		if i >= 10 {
		//			break
		//		}
		t.Logf("%+v\n", v)
	}
	t.Log(len(em.Metrics))
}

func TestRegistrationMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/TileMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\TileMetricsOut.bin`
	filename := `\\ussd-prd-isi04\Voyager\160401_SN924_2689_A057KBAAVY\InterOp\RegistrationMetricsOut.bin`
	em := RegistrationMetricsInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log(em.Version, em.SSize, em.NumOfChannels, em.NumberOfSubRegions)
	for i, v := range em.Metrics {
		if i >= 10 {
			break
		}
		t.Log(v.LTC)
		t.Logf("%+v\n", v.Channels)
	}
	t.Log(len(em.Metrics))
}

func xTestFwhmGridMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/ExtractionMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\ExtractionMetricsOut.bin`
	//	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\ExtractionMetricsOut.bin`
	//	filename := `\\ussd-prd-isi04\Voyager\160106_HWI-ST347_1760_BHKFGGCCXX\InterOp\FwhmGridMetricsOut.bin`
	filename := `\\ussd-prd-isi04\Voyager\iPeg\150618_G134_0342_AH2YWGBBXX\InterOp\FwhmGridMetricsOut.bin`
	pf := FwhmMetricsInfo{Filename: filename}
	err := pf.Parse()
	if err != nil {
		t.Fatal(err.Error())
	}

	for i, m := range pf.Metrics {
		t.Log(m.LaneNum, m.TileNum, m.Cycle)

		for _, c := range m.Channels {
			t.Logf("%d, %d ", c.Channel, len(c.Fwhm))
			//			for idx, f := range c.Fwhm {
			//				t.Log(idx, f)
			//			}
		}
		//		t.Log("\n")
		//		for _, v := range m.PFCluster {
		//			t.Logf("%d, ", v)
		//		}
		//		t.Log("\n")
		if i >= 10 {
			break
		}
	}
	t.Logf("%+v\n", pf)

}

func TestPFGridMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/ExtractionMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\ExtractionMetricsOut.bin`
	//	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\ExtractionMetricsOut.bin`
	filename := `\\ussd-prd-isi04\Voyager\160106_HWI-ST347_1760_BHKFGGCCXX\InterOp\PFGridMetricsOut.bin`
	pf := PFMetricsInfo{Filename: filename}
	err := pf.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", pf)

	for i, m := range pf.Metrics {
		t.Log(m.LaneNum, m.TileNum)

		//		for _, v := range m.RawCluster {
		//			t.Logf("%d, ", v)
		//		}
		//		t.Log("\n")
		//		for _, v := range m.PFCluster {
		//			t.Logf("%d, ", v)
		//		}
		//		t.Log("\n")
		if i >= 10 {
			break
		}
	}
}

func TestExtractionMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/ExtractionMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\ExtractionMetricsOut.bin`
	//	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\ExtractionMetricsOut.bin`
	//	filename := `\\ussd-prd-isi04\ARG\110929_GAIIX-596_00089_FC70LYG_WC_exo+\InterOp\ExtractionMetricsOut.bin`
	filename := `\\ussd-prd-isi04\voyager\161026_VP2-06_0068_AH5LWDMCVY\InterOp\C10.1\ExtractionMetricsOut.bin`
	em := ExtractionInfo{Filename: filename}
	err := em.Parse3(filename)
	if err != nil {
		t.Error(err)
	}
	t.Log("Max Cycle", em.GetMaxCycle())
	t.Log("Last CIF TIME", GetTime(em.GetLatestCIFTime()))
	for i, m := range em.Metrics {
		t.Log(m.LaneNum, m.TileNum, m.Cycle)
		if i >= 1 {
			break
		}
	}
}

//C25.1/TileMetricsOut.bin
func xTestTileMetricsRTA3(t *testing.T) {
	filename := `\\ussd-prd-isi04\Voyager\161026_VP2-06_0068_AH5LWDMCVY\InterOp\C25.1\TileMetricsOut.bin`
	em := TileInfo{Filename: filename}
	err := em.ParseRTA3()
	if err != nil {
		t.Error(err)
	}
	for i, v := range em.Metrics3 {
		if i >= 1000 {
			break
		}
		t.Logf("%+v\n", v)
	}
	t.Log(len(em.Metrics3))
}

func xTestTileMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/TileMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\TileMetricsOut.bin`
	//	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\TileMetricsOut.bin`
	filename := `\\ussd-prd-isi04\Voyager\160701_VP1-08_0164_A027BCABVY\InterOp\TileMetricsOut.bin`
	em := TileInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	for i, v := range em.Metrics {
		if i >= 1000 {
			break
		}
		t.Log(v)
	}
	t.Log(len(em.Metrics))
}

func xTestQMetrics7(t *testing.T) {
	//	filename := "./test_data/InterOp/QMetricsOut.bin"
	//	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\QMetricsOut.bin`
	//	filename := `\\ussd-prd-isi04\Voyager\160701_VP1-08_0164_A027BCABVY\InterOp\QMetricsOut.bin`
	filename := `\\ussd-prd-isi04\Voyager\161026_VP2-06_0068_AH5LWDMCVY\InterOp\C25.1\QMetricsOut.bin`
	em := QMetricsInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	for i, v := range em.Metrics7 {
		if i >= 100 {
			break
		}
		t.Log(v)
	}
	t.Log(em.EnableQbin)
}

func TestQMetrics_version5(t *testing.T) {
	//	filename := "./test_data/InterOp/QMetricsOut_version5.bin"
	filename := "./test_data/InterOp/QMetricsOut_version5.bin"
	em := QMetricsInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log(em.EnableQbin, em.NumQscores)
	//	for i, v := range em.Metrics {
	//		t.Log(v)
	//		if i >= 10 {
	//			break
	//		}
	//	}
	if !em.EnableQbin {
		t.Logf("unable to parse qbin-ed Qmetrics")
	}

}

func xTestQMetrics_version6(t *testing.T) {
	//	filename := "./test_data/InterOp/QMetricsOut_version6.bin"
	//	filename := "./test_data/QMetricsOut.bin"
	filename := `\\sd-isilon\trex\Opus\150512_ST-E00107_0505_BH01V5CFXX\InterOp\QMetricsOut.bin`
	em := QMetricsInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}

	t.Log(em.Version, em.EnableQbin, em.NumQscores, em.QbinConfig.ReMapScores)
	if !em.EnableQbin {
		t.Logf("unable to parse qbin-ed Qmetrics")
	}
	if em.Error() != "" {
		t.Errorf(em.Error())
	}
}

func TestErrorMetrics4(t *testing.T) {
	//	filename := "./test_data/InterOp/ErrorMetricsOut.bin"
	filename := `\\ussd-prd-isi04\Voyager\161026_VP2-06_0068_AH5LWDMCVY\InterOp\C25.1\ErrorMetricsOut.bin `
	em := ErrorInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log(em.Version, em.SSize)
	t.Log(len(em.Metrics4))

	dim := em.GetDimMax()
	t.Logf("%+v\n", dim)
	tileER := em.ErrorRateByTile(nil)
	t.Logf("%+v\n", tileER)
	for _, ln := range tileER.Lanes {
		//		t.Logf("%+v\n", ln)
		b, err := json.Marshal(ln)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%s\n", string(b))
	}

	exMap := make(map[uint16]bool)
	for i := uint16(150); i < uint16(301); i++ {
		exMap[i] = true
	}
	bubbles := em.BubbleCounter(exMap)
	t.Logf(" bubbles\n %+v\n", bubbles)
	for _, ln := range bubbles.Lanes {
		//		t.Logf("%+v\n", ln)
		b, err := json.Marshal(ln)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%s\n", string(b))
	}
	bubSum := bubbles.GetBubbleSum(nil)
	t.Logf(" bubbles Sumamry\n %+v\n", bubSum)
	return
	for i, each := range em.Metrics {
		if i >= 1000 {
			break
		}
		t.Log(each.TileNum, each.LaneNum)
	}
}

func xTestErrorMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/ErrorMetricsOut.bin"
	filename := `\\ussd-prd-isi04\Voyager\Voychem\160212_E355_0238_A00WNCAAVY\InterOp\ErrorMetricsOut.bin`
	em := ErrorInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}

	t.Log(len(em.Metrics))
	dim := em.GetDimMax()
	t.Logf("%+v\n", dim)
	tileER := em.ErrorRateByTile(nil)
	t.Logf("%+v\n", tileER)
	for _, ln := range tileER.Lanes {
		//		t.Logf("%+v\n", ln)
		b, err := json.Marshal(ln)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%s\n", string(b))
	}
	exMap := make(map[uint16]bool)
	for i := uint16(150); i < uint16(301); i++ {
		exMap[i] = true
	}
	bubbles := em.BubbleCounter(exMap)
	t.Logf(" bubbles\n %+v\n", bubbles)
	for _, ln := range bubbles.Lanes {
		//		t.Logf("%+v\n", ln)
		b, err := json.Marshal(ln)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%s\n", string(b))
	}
	bubSum := bubbles.GetBubbleSum(nil)
	t.Logf(" bubbles Sumamry\n %+v\n", bubSum)
	return
	for i, each := range em.Metrics {
		if i >= 1000 {
			break
		}
		t.Log(each.TileNum, each.LaneNum)
	}
}

func TestCorrectedIntMetrics(t *testing.T) {
	filename := "./test_data/InterOp/CorrectedIntMetricsOut.bin"
	em := CorrectIntInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}

	t.Log(len(em.Metrics))
}

func TestImageMetrics(t *testing.T) {
	//version 2
	filename := `\\ussd-prd-isi04\Voyager\160401_SN924_2689_A057KBAAVY\InterOp\ImageMetricsOut.bin`
	em := ImageInfo{Filename: filename}
	err := em.Parse()
	t.Log(em.Version)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))

	//version 1
	filename = "./test_data/InterOp/ImageMetricsOut.bin"
	em = ImageInfo{Filename: filename}
	err = em.Parse()
	t.Log(em.Version)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))

}

func TestIndexMetrics(t *testing.T) {
	filename := "./test_data/InterOp/IndexMetricsOut.bin"
	em := IndexInfo{Filename: filename}
	err := em.Parse()
	_ = err
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))
}

func TestControlMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ControlMetricsOut.bin"
	em := ControlInfo{Filename: filename}
	err := em.Parse()
	_ = err
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))
}

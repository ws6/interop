package interop

import (
	"testing"
)

func TestExtractionMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/ExtractionMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\ExtractionMetricsOut.bin`
	//	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\ExtractionMetricsOut.bin`
	filename := `\\ussd-prd-isi04\ARG\110929_GAIIX-596_00089_FC70LYG_WC_exo+\InterOp\ExtractionMetricsOut.bin`
	em := ExtractionInfo{Filename: filename}
	err := em.Parse()
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

func TestTileMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/TileMetricsOut.bin"
	//	filename := `\\ussd-prd-isi04\Voyager\150910_E360_0084_AHG75WCCXX\InterOp\TileMetricsOut.bin`
	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\TileMetricsOut.bin`
	em := TileInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	for i, v := range em.Metrics {
		if i >= 100 {
			break
		}
		t.Log(v)
	}
	t.Log(len(em.Metrics))
}

func TestQMetrics(t *testing.T) {
	//	filename := "./test_data/InterOp/QMetricsOut.bin"
	filename := `C:\Users\jliu1\GolangProjects\src\github.com\ws6\raptor\test_data\data\flowcells\150924_GAIIX-778_00444_FC66GDFAAXX\InterOp\QMetricsOut.bin`
	em := QMetricsInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	//	for _, v := range em.Metrics {
	//		t.Log(v)
	//	}
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

func TestQMetrics_version6(t *testing.T) {
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

func TestErrorMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ErrorMetricsOut.bin"
	em := ErrorInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}

	t.Log(len(em.Metrics))
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
	filename := "./test_data/InterOp/ImageMetricsOut.bin"
	em := ImageInfo{Filename: filename}
	err := em.Parse()
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

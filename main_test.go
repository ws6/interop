package interop

import (
	"testing"
)

func TestExtractionMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ExtractionMetricsOut.bin"
	em := ExtractionInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log("Max Cycle", em.GetMaxCycle())
	t.Log("Last CIF TIME", GetTime(em.GetLatestCIFTime()))
}

func TestTileMetrics(t *testing.T) {
	filename := "./test_data/InterOp/TileMetricsOut.bin"
	em := TileInfo{Filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	//	for _, v := range em.Metrics {
	//		t.Log(v)
	//	}
	t.Log(len(em.Metrics))
}

func TestQMetrics(t *testing.T) {
	filename := "./test_data/InterOp/QMetricsOut.bin"
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

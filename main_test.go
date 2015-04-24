package interop

import (
	"testing"
)

func TestExtractionMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ExtractionMetricsOut.bin"
	em := ExtractionInfo{filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log("Max Cycle", em.GetMaxCycle())
	t.Log("Last CIF TIME", GetTime(em.GetLatestCIFTime()))
}

func TestTileMetrics(t *testing.T) {
	filename := "./test_data/InterOp/TileMetricsOut.bin"
	em := TileInfo{filename: filename}
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
	em := QMetricsInfo{filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	//	for _, v := range em.Metrics {
	//		t.Log(v)
	//	}
	t.Log(em.EnableQbin)
}

func TestErrorMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ErrorMetricsOut.bin"
	em := ErrorInfo{filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}

	t.Log(len(em.Metrics))
}

func TestCorrectedIntMetrics(t *testing.T) {
	filename := "./test_data/InterOp/CorrectedIntMetricsOut.bin"
	em := CorrectIntInfo{filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}

	t.Log(len(em.Metrics))
}

func TestImageMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ImageMetricsOut.bin"
	em := ImageInfo{filename: filename}
	err := em.Parse()
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))
}

func TestIndexMetrics(t *testing.T) {
	filename := "./test_data/InterOp/IndexMetricsOut.bin"
	em := IndexInfo{filename: filename}
	err := em.Parse()
	_ = err
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))
}

func TestControlMetrics(t *testing.T) {
	filename := "./test_data/InterOp/ControlMetricsOut.bin"
	em := ControlInfo{filename: filename}
	err := em.Parse()
	_ = err
	if err != nil {
		t.Error(err)
	}
	t.Log(len(em.Metrics))
}

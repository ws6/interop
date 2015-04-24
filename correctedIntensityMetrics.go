package interop

import (
	"encoding/binary"
	"os"
)

type CorrectIntMetrics struct {
	LaneNum         uint16
	TileNum         uint16
	Cycle           uint16
	AvgIntensity    uint16
	Avg_Int_A       uint16 //Average Corrected Intensity For Channel A
	Avg_Int_C       uint16
	Avg_Int_G       uint16
	Avg_Int_T       uint16
	Avg_Called_A    uint16 //Average Corrected Intensity for called cluster of Base A
	Avg_Called_C    uint16
	Avg_Called_G    uint16
	Avg_Called_T    uint16
	BaseCall_NoCall float32 //number of base calls (float) for No Call
	BaseCall_A      float32 //number of base calls (float) for Channel A
	BaseCall_C      float32
	BaseCall_G      float32
	BaseCall_T      float32
	NoiseRatio      float32 //signal to noise ratio
}

type CorrectIntInfo struct {
	filename string
	Version  uint8
	SSize    uint8
	Metrics  []*CorrectIntMetrics
	err      error
}

func (self *CorrectIntInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	header, err := GetHeader(file)

	if err != nil {
		self.err = err
		return self.err
	}
	self.Version = header.Version
	self.SSize = header.SSize

	for {
		em := new(CorrectIntMetrics)
		err = binary.Read(header.Buf, binary.LittleEndian, em)
		if err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			break
		}

		self.Metrics = append(self.Metrics, em)
	}
	return self.err
}

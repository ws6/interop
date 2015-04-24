package interop

import (
	"encoding/binary"
	"os"
)

type ErrorMetrics struct {
	LaneNum         uint16
	TileNum         uint16
	Cycle           uint16
	ErrorRate       float32
	NumPerfectReads uint32
	Num_1_Error     uint32
	Num_2_Error     uint32
	Num_3_Error     uint32
	Num_4_Error     uint32
}

type ErrorInfo struct {
	filename string
	Version  uint8
	SSize    uint8
	Metrics  []*ErrorMetrics
	err      error
}

func (self *ErrorInfo) Parse() error {
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
		em := new(ErrorMetrics)
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

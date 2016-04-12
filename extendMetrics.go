package interop

import (
	"bufio"
	"encoding/binary"
	"os"
)

//byte 0: file version number (1)
//byte 1: length of each record
//bytes (N * 10 + 2)  - (N *10 + 11): record:
//        2 bytes: lane number (uint16)
//        2 bytes: tile number (uint16)
//        2 bytes: metric code (uint16)
//        4 bytes: metric value (float)
//where N is the record index and possible metric codes are:
//        code 0: number of clusters occupying wells on a patterned flowcell

var (
	CLUSTER_OCCUPIED = uint16(0)
)

type ExtendMetrics struct {
	LaneNum uint16
	TileNum uint16
	Code    uint16
	Value   float32
}

type ExtendMetricsInfo struct {
	Filename string
	Version  uint8
	SSize    uint8
	Metrics  []*ExtendMetrics
	err      error
}

func (self *ExtendMetricsInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.Filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	buffer := bufio.NewReader(file)

	//read version
	if err := binary.Read(buffer, binary.LittleEndian, &self.Version); err != nil {
		return err
	}

	//read length of each record
	if err := binary.Read(buffer, binary.LittleEndian, &self.SSize); err != nil {
		return err
	}

	for {
		em := new(ExtendMetrics)

		if err = binary.Read(buffer, binary.LittleEndian, em); err != nil {
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

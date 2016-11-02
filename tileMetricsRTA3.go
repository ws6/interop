package interop

import (
	"encoding/binary"
	"fmt"
	"io"

	"os"
)

//Format for version 3:
//byte 0: file version number (2)
//byte 1: length of each record (15)
//bytes 2-5: area of a tile in mm^2 (float)
//bytes (N * 15 + 6)  - (N * 15 + 20): record:
//        2 bytes: lane number (uint16)
//        4 bytes: tile number (uint32)
//        1 byte: metricCode, the metric code (char)
//        if(metricCode == 't')
//               4 bytes: cluster count (float)
//               4 bytes: PF cluster count (float)
//        else if(metricCode == 'r')
//               4 bytes: read number (uint32)
//               4 bytes: % aligned (float)
//        else if(metricCode == '\0')
//                8 bytes: 0
//where N is the record index
type Cluster struct {
	ClusterCount   float32 //if  MetricCode == 't'
	PFClusterCount float32 //if  MetricCode == 't'
}

type ReadAlignment struct {
	NumberRead uint32
	PctAligned float32
}

type Padding struct {
	P1 uint32
	P2 uint32
}

type LT struct {
	LaneNum uint16
	TileNum uint32
}

type TileMetrics3 struct {
	LT
	MetricCode byte
	Cluster
	ReadAlignment
	//	Padding
}

func (self *TileInfo) ParseRTA3() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.Filename)
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
	if self.Version != 3 {
		return fmt.Errorf("Not RTA version 3, got %d", self.Version)
	}
	self.SSize = header.SSize

	if err := binary.Read(header.Buf, binary.LittleEndian, &self.AreaSize); err != nil {
		return err
	}
	//	fmt.Println(self.Version, self.SSize, self.AreaSize)
	for {
		em := new(TileMetrics3)
		if err := binary.Read(header.Buf, binary.LittleEndian, &em.LT); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		//read metricsCode
		if err := binary.Read(header.Buf, binary.LittleEndian, &em.MetricCode); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		//		p := Padding{}
		//		if err := binary.Read(header.Buf, binary.LittleEndian, &p); err != nil {
		//			if err == io.EOF {
		//				return nil
		//			}
		//			return err
		//		}
		//		fmt.Println(string(em.MetricCode), em.MetricCode, em.LT)
		//		continue
		if em.MetricCode == 't' {
			if err := binary.Read(header.Buf, binary.LittleEndian, &em.Cluster); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
		if em.MetricCode == 'r' {
			if err := binary.Read(header.Buf, binary.LittleEndian, &em.ReadAlignment); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
		if em.MetricCode == 0 {
			p := Padding{}
			if err := binary.Read(header.Buf, binary.LittleEndian, &p); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
		self.Metrics3 = append(self.Metrics3, em)
	}
	return nil
}

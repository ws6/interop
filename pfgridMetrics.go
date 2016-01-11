package interop

//PF Grid Metrics (PFGridMetricsOut.bin) – INTERNAL USE ONLY
import (
	"bufio"
	"encoding/binary"
	//	"fmt"
	//	"math"
	"os"
)

type PFSubTileMetrics struct {
	LaneNum    uint16
	TileNum    uint16
	RawCluster []uint32
	PFCluster  []uint32
}
type PFMetricsInfo struct {
	Filename string
	Version  uint8
	SSize    uint16
	NumX     uint16
	NumY     uint16
	BinArea  float32
	Metrics  []*PFSubTileMetrics
	err      error
}

func (self *PFMetricsInfo) Parse() error {
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

	//read subtile number of X a.k.a columns
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumX); err != nil {
		return err
	}
	//read subtile number of Y a.k.a rows
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumY); err != nil {
		return err
	}

	//read subtile area size in mm2
	if err := binary.Read(buffer, binary.LittleEndian, &self.BinArea); err != nil {
		return err
	}
	numSubTiles := self.NumX * self.NumY
	for {
		//read RawClusters for all subtiles
		//from x to y
		//e,g : x=4 y =2
		//[1,1] [1,2] [1,3] [1,4]
		//[2,1] [2,2] [2,3] [2,4]
		m := new(PFSubTileMetrics)
		//read lane number
		if err := binary.Read(buffer, binary.LittleEndian, &m.LaneNum); err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
		//read tile number
		if err := binary.Read(buffer, binary.LittleEndian, &m.TileNum); err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
		//read Raw Cluster

		for i := uint16(0); i < numSubTiles; i++ {
			var f uint32
			if err := binary.Read(buffer, binary.LittleEndian, &f); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				return err
			}
			m.RawCluster = append(m.RawCluster, f)
		}
		//read for PF clusters
		for i := uint16(0); i < numSubTiles; i++ {
			var f uint32
			if err := binary.Read(buffer, binary.LittleEndian, &f); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				return err
			}
			m.PFCluster = append(m.PFCluster, f)
		}
		self.Metrics = append(self.Metrics, m)
	}
	return err
}

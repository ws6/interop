package interop

//PF Grid Metrics (PFGridMetricsOut.bin) – INTERNAL USE ONLY
import (
	"bufio"
	"encoding/binary"
	//	"fmt"
	//	"math"
	"os"
)

type FwhmChannel struct {
	Channel uint8
	Fwhm    []float32
}

type FwhmSubTileMetrics struct {
	LTC
	//	ChannelIndex uint8
	Channels []*FwhmChannel
}
type FwhmMetricsInfo struct {
	Filename    string
	Version     uint8
	NumX        uint8
	NumY        uint8
	NumChannels uint8
	SSize       uint16

	Metrics []*FwhmSubTileMetrics
	err     error
}

func (self *FwhmMetricsInfo) Parse() error {
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

	//read subtile number of X a.k.a columns
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumX); err != nil {
		return err
	}
	//read subtile number of Y a.k.a rows
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumY); err != nil {
		return err
	}

	//read   NumChannels
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumChannels); err != nil {
		return err
	}

	//read length of each record
	if err := binary.Read(buffer, binary.LittleEndian, &self.SSize); err != nil {
		return err
	}
	//overflowce for uint8 if numSubTiles is greater than 256
	numSubTiles := int(self.NumX) * int(self.NumY)
	for {
		//read RawClusters for all subtiles
		//from x to y
		//e,g : x=4 y =2
		//[1,1] [1,2] [1,3] [1,4]
		//[2,1] [2,2] [2,3] [2,4]
		m := new(FwhmSubTileMetrics)
		//read lane number
		if err := binary.Read(buffer, binary.LittleEndian, &m.LTC); err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
		//		for j := uint8(0); j < self.NumChannels; j++ {
		for j := uint8(0); j < uint8(4); j++ { //always four channels
			fwhmChannel := new(FwhmChannel)
			fwhmChannel.Channel = j
			for i := 0; i < numSubTiles; i++ {
				var f float32
				if err := binary.Read(buffer, binary.LittleEndian, &f); err != nil {
					self.err = err
					if err.Error() == "EOF" {
						return nil
					}
					return err
				}
				fwhmChannel.Fwhm = append(fwhmChannel.Fwhm, f)
			}
			m.Channels = append(m.Channels, fwhmChannel)
		}
		self.Metrics = append(self.Metrics, m)
	}
	return err
}

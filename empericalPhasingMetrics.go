package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	//	"math"
	"io"
	"os"
)

type PhasingMetrics struct {
	LTC
	Phasing    float32
	PrePhasing float32
}

type EmpericalPhasingInfo struct {
	Filename string
	Version  uint8
	SSize    uint8
	Metrics  []*PhasingMetrics
}

func (self *EmpericalPhasingInfo) Parse() error {
	file, err := os.Open(self.Filename)
	if err != nil {
		return err
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
		m := new(PhasingMetrics)
		//read LTC
		if err := binary.Read(buffer, binary.LittleEndian, m); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("LTC err:%s", err)
		}
		self.Metrics = append(self.Metrics, m)
	}
	return nil
}

func (self *EmpericalPhasingInfo) FilterByTileMap(tm *[]LaneTile) *EmpericalPhasingInfo {
	ret := &EmpericalPhasingInfo{
		Filename: self.Filename,
		Version:  self.Version,
		SSize:    self.SSize,
	}

	tmap := MakeLaneTileMap(tm)
	ret.Metrics = make([]*PhasingMetrics, 0)
	for _, t := range self.Metrics {
		if _, ok := tmap[t.LaneNum]; !ok {
			continue
		}
		if use, ok := tmap[t.LaneNum][t.TileNum]; !ok || !use {
			continue
		}
		//!!! only use ref
		ret.Metrics = append(ret.Metrics, t)
	}
	return nil
}

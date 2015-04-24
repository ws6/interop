package interop

import (
	"bufio"
	"encoding/binary"
	"os"
)

type HeaderInfo struct {
	Version uint8
	SSize   uint8
	Buf     *bufio.Reader
}

func GetHeader(file *os.File) (header *HeaderInfo, err error) {
	header = new(HeaderInfo)

	//	defer file.Close()

	header.Buf = bufio.NewReader(file)

	err = binary.Read(header.Buf, binary.LittleEndian, &header.Version)
	if err != nil {
		return
	}
	err = binary.Read(header.Buf, binary.LittleEndian, &header.SSize)
	if err != nil {
		return
	}
	return
}

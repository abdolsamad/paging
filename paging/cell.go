package paging

import (
	"fmt"
	"pager/serializarion"
)

type Cell []byte

type CellComparator func(cell1 Cell, cell2 Cell) bool

func NewCell(payload []byte) (Cell, error) {
	cellSize := len(payload)
	if cellSize > PAGE_SIZE-PAGE_HEADER_SIZE {
		return nil, fmt.Errorf("too large payload")
	}
	cellSizeBytes := serializarion.SerializeUint16(uint16(cellSize))
	cellBytes := append(cellSizeBytes, payload...)
	return cellBytes, nil
}
func (c Cell) GetSize() uint16 {
	return c.GetPayloadSize() + 2
}
func (c Cell) GetPayloadSize() uint16 {
	return serializarion.DeserializeUint16(c[:2])
}
func (c Cell) GetPayload() []byte {
	return c[2:]
}

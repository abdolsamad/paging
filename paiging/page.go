package paiging

import (
	"fmt"
	"pager/serializarion"
)

type Page []byte

func NewPage() Page {
	page := make([]byte, PAGE_SIZE)
	pageHeader := NewRawPageHeader(PAGE_FLAG_LEAF)
	copy(page[:PAGE_HEADER_SIZE], pageHeader)
	return page
}

func (p Page) GetRawHeader() RawPageHeader {
	return RawPageHeader(p[:PAGE_HEADER_SIZE])
}

func (p Page) GetHeader() *PageHeader {
	return &PageHeader{
		Flags:                p.GetRawHeader().GetFlag(),
		CellsCount:           p.GetRawHeader().GetCellsCount(),
		FirstCellOffset:      p.GetRawHeader().GetFirstCellOffset(),
		FirstFreeBlockOffset: p.GetRawHeader().GetFirstFreeBlockOffset(),
		FragmentedBytesCount: p.GetRawHeader().GetFragmentedBytesCount(),
	}
}

func (p Page) GetFreeBytes() uint16 {
	var totalSize uint16 = PAGE_SIZE
	var pageHeaderSize uint16 = PAGE_HEADER_SIZE
	var cellOffsetsListSize uint16 = 2 * p.GetRawHeader().GetCellsCount()
	var bytesUsedByContentArea uint16 = PAGE_SIZE - p.GetRawHeader().GetFirstCellOffset() - 1

	return totalSize - pageHeaderSize - cellOffsetsListSize - bytesUsedByContentArea
}

func (p Page) GetCellsOffsets() []uint16 {
	cellsCount := p.GetRawHeader().GetCellsCount()
	cellOffsets := make([]uint16, cellsCount)
	for i := uint16(0); i < cellsCount; i++ {
		cellOffsets[i] = serializarion.DeserializeUint16(p[PAGE_HEADER_SIZE+i : PAGE_HEADER_SIZE+i+2])
	}
	return cellOffsets
}

func (p Page) AddCell(cell Cell, comparator CellComparator) (uint16, error) { //returns the offset of created cell
	if cell.GetSize()+2 > p.GetFreeBytes() {                                  // 2 is for offset entry
		return 0, fmt.Errorf("not enough space in the page")
	}
	cellLocation := p.GetRawHeader().GetFirstFreeBlockOffset() - cell.GetSize() + 1
	copy(p[cellLocation:], cell)
	//update First Free Block Offset
	p.GetRawHeader().SetFirstFreeBlockOffset(cellLocation - 1)
	//update cell counts
	p.addCellOffsetToList(cell, cellLocation, comparator) //TODO:
	p.GetRawHeader().IncreaseCellsCount()
	return cellLocation, nil
}

func (p Page) getCellOffsetAt(index uint16) uint16 {
	offsetBytes := p[PAGE_HEADER_SIZE+2*index : PAGE_HEADER_SIZE+2*index+2]
	return serializarion.DeserializeUint16(offsetBytes)
}

func (p Page) GetCellAt(index uint16) Cell {
	cellOffset := p.getCellOffsetAt(index)
	payloadSize := serializarion.DeserializeUint16(p[cellOffset : cellOffset+2])
	cellPayload := p[cellOffset+2 : cellOffset+2+payloadSize]
	cell, _ := NewCell(cellPayload)
	return cell
}

func (p Page) addCellOffsetToList(cell Cell, cellLocation uint16, comparator CellComparator) {
	cellOffsets := p.GetCellsOffsets()
	cellsCount := uint16(len(cellOffsets))
	insertLocation := uint16(0)
	for i := uint16(0); i < cellsCount; i++ {
		tempCell := p.GetCellAt(i)
		if comparator(cell, tempCell) {
			insertLocation = i
			break
		}
	}

	newCellOffsets := append(cellOffsets, 0)
	copy(newCellOffsets[insertLocation+1:], newCellOffsets[insertLocation:])
	newCellOffsets[insertLocation] = cellLocation
	p.replaceCellOffsets(newCellOffsets)
}

func (p Page) replaceCellOffsets(offsets []uint16) {
	offsetArrLength := uint16(len(offsets))
	for i := uint16(0); i < offsetArrLength; i++ {
		byted := serializarion.SerializeUint16(offsets[i])
		p[PAGE_HEADER_SIZE+2*i] = byted[0]
		p[PAGE_HEADER_SIZE+2*i+1] = byted[1]
	}
}

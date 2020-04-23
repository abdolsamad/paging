package paging

import (
	"fmt"

	"github.com/abdolsamad/paging/serializarion"
)

const InvalidOffset = uint16(0)

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

func (p Page) AddCell(cell Cell, comparator CellComparator) (uint16, error) { //returns the offset of created cell
	if cell.GetSize()+2 > p.GetFreeBytes() { // 2 is for offset entry
		return 0, fmt.Errorf("not enough space in the page")
	}
	var cellOffset uint16
	freeBlockObject := p.allocateBlockInEmptyAreas(cell.GetSize())
	if freeBlockObject != nil {
		cellOffset = p.putCellInFreeBlockUnchecked(cell, freeBlockObject)
	} else {
		cellOffset = p.allocateNewBlock(cell.GetSize())
		//update First Cell Offset
		p.GetRawHeader().SetFirstCellOffset(cellOffset)
	}
	copy(p[cellOffset:cellOffset+cell.GetSize()], cell)
	//update cell offsets list
	p.addCellOffsetToList(cell, cellOffset, comparator) //TODO:
	//update cell counts
	p.GetRawHeader().IncreaseCellsCount()
	return cellOffset, nil
}

func (p Page) DeleteCell(offset uint16) error {
	cell, err := p.GetCellAtOffset(offset)
	if err != nil {
		return err
	}
	cellSize := cell.GetSize()
	cellEndOffset := offset + cellSize
	rawHeader := p.GetRawHeader()
	cellOffsetBefore := p.getCellOffsetJustBefore(offset)
	freeBlockBefore := p.getFreeBlockJustBefore(offset)
	cellOffsetAfter := p.getCellOffsetJustAfter(offset)
	freeBlockAfter := p.getFreeBlockJustAfter(offset)

	if cellOffsetBefore == InvalidOffset { //first cell
		if cellOffsetAfter == InvalidOffset { //only cell
			rawHeader.SetFirstCellOffset(InvalidOffset)
			rawHeader.SetFirstFreeBlockOffset(InvalidOffset)
		} else {
			rawHeader.SetFirstCellOffset(cellOffsetAfter)
			if freeBlockAfter != nil { // a free block exist after removed cell
				if freeBlockAfter.Offset > cellOffsetAfter { //a cell and after that a free block
					rawHeader.SetFirstFreeBlockOffset(freeBlockAfter.Offset)
				} else { //a free block and a cell after that
					//free block after removed cell will be removed. the free block after that(if any) would be the first free block
					next := freeBlockAfter.GetNext()
					if next != nil {
						rawHeader.SetFirstFreeBlockOffset(next.Offset)
					}
				}
			} else { //no free block was after deleted cell. it can't be any before that(because it was the first cell), so there is'nt any free block at all
				rawHeader.SetFirstFreeBlockOffset(InvalidOffset)
			}
		}
	} else {
		isCellBefore := freeBlockBefore == nil || freeBlockBefore.Offset > cellOffsetBefore
		isCellAfter := cellOffsetAfter != InvalidOffset &&
			(freeBlockAfter == nil || freeBlockAfter.Offset > cellOffsetAfter)
		isNothingAfter := cellOffsetAfter == InvalidOffset && freeBlockAfter == nil
		isFreeBefore := freeBlockBefore != nil && freeBlockBefore.Offset < cellOffsetBefore
		isFreeAfter := !isNothingAfter && !isCellAfter

		if isCellBefore && isCellAfter { //<cell>.?<THIS>.?<cell>
			previousCellEndOffset := cellOffsetBefore + p.getCellAtOffsetUnchecked(cellOffsetBefore).GetSize()
			removedFragmentedBytes := offset - previousCellEndOffset
			removedFragmentedBytes += cellOffsetAfter - cellEndOffset
			rawHeader.DecreaseFragmentedBytesCount(byte(removedFragmentedBytes))
			if freeBlockBefore != nil {
				p.insertFreeBlockUnchecked(previousCellEndOffset, cellOffsetAfter-previousCellEndOffset, freeBlockBefore.Offset)
			} else {
				p.insertFreeBlockUnchecked(previousCellEndOffset, cellOffsetAfter-previousCellEndOffset, InvalidOffset)
			}
		} else if isFreeBefore && isCellAfter { //<free><THIS>.?<cell>
			removedFragmentedBytes := cellOffsetAfter - cellEndOffset
			rawHeader.DecreaseFragmentedBytesCount(byte(removedFragmentedBytes))
			freeBlockBefore.header.setBlockSize(cellOffsetAfter - freeBlockBefore.Offset)
		} else if isFreeBefore && isFreeAfter { //<free><THIS><free>
			newSize := freeBlockBefore.GetSize() + freeBlockAfter.GetSize() + cellSize
			freeBlockBefore.header.setNextFreeBlockOffset(freeBlockAfter.header.getNextFreeBlockOffset())
			freeBlockBefore.SetSize(newSize)
		} else if isCellBefore && isFreeAfter { //<cell>.?<THIS><free>
			previousCellEndOffset := cellOffsetBefore + p.getCellAtOffsetUnchecked(cellOffsetBefore).GetSize()
			removedFragmentedBytes := offset - previousCellEndOffset
			rawHeader.DecreaseFragmentedBytesCount(byte(removedFragmentedBytes))
			freeBlockAfter.header.setBlockSize(freeBlockAfter.Offset - previousCellEndOffset)
			copy(p[previousCellEndOffset:previousCellEndOffset+FREE_BLOCK_HEADER_SIZE],
				p[freeBlockAfter.Offset:freeBlockAfter.Offset+FREE_BLOCK_HEADER_SIZE])
			if freeBlockBefore != nil {
				freeBlockBefore.header.setNextFreeBlockOffset(previousCellEndOffset)
			} else {
				rawHeader.SetFirstFreeBlockOffset(previousCellEndOffset)
			}
		} else if isCellBefore && isNothingAfter { //<cell>.?<THIS>.?<END>
			previousCellEndOffset := cellOffsetBefore + p.getCellAtOffsetUnchecked(cellOffsetBefore).GetSize()
			removedFragmentedBytes := offset - previousCellEndOffset
			removedFragmentedBytes += PAGE_SIZE - cellEndOffset
			rawHeader.DecreaseFragmentedBytesCount(byte(removedFragmentedBytes))
			if freeBlockBefore != nil {
				p.insertFreeBlockUnchecked(previousCellEndOffset, PAGE_SIZE-previousCellEndOffset, freeBlockBefore.Offset)
			} else {
				p.insertFreeBlockUnchecked(previousCellEndOffset, PAGE_SIZE-previousCellEndOffset, InvalidOffset)
			}
		} else if isCellBefore && isNothingAfter { //<free><THIS>.?<END>
			removedFragmentedBytes := PAGE_SIZE - cellEndOffset
			rawHeader.DecreaseFragmentedBytesCount(byte(removedFragmentedBytes))
			freeBlockBefore.SetSize(PAGE_SIZE - freeBlockBefore.Offset)
		} else {
			panic("What the hell? this shouldn't happened at all!!")
		}
	}

	p.deleteCellOffsetFromList(offset)
	rawHeader.DecreaseCellsCount()
	return nil
}

func (p Page) getCellOffsetJustBefore(offset uint16) uint16 {
	result := uint16(0)
	for i := uint16(0); i < p.GetRawHeader().GetCellsCount(); i++ {
		offsetOfCell := p.getNthCellOffsetUnchecked(i)
		if offsetOfCell < offset && offsetOfCell > result {
			result = offsetOfCell
		}
	}
	return result
}

func (p Page) getCellJustBefore(offset uint16) Cell {
	cellOffset := p.getCellOffsetJustBefore(offset)
	if cellOffset == 0 {
		return nil
	} else {
		return p.getCellAtOffsetUnchecked(cellOffset)
	}
}

func (p Page) getCellOffsetJustAfter(offset uint16) uint16 {
	result := InvalidOffset
	for i := uint16(0); i < p.GetRawHeader().GetCellsCount(); i++ {
		offsetOfCell := p.getNthCellOffsetUnchecked(i)
		if offsetOfCell > offset && (offsetOfCell < result || result == InvalidOffset) {
			result = offsetOfCell
		}
	}
	return result
}

func (p Page) getCellJustAfter(offset uint16) Cell {
	cellOffset := p.getCellOffsetJustAfter(offset)
	if cellOffset == 0 {
		return nil
	} else {
		return p.getCellAtOffsetUnchecked(cellOffset)
	}
}

func (p Page) getFreeBlockJustBefore(offset uint16) *freeBlock {
	current := p.getFirstFreeBlock()
	if current == nil || current.Offset > offset {
		return nil
	}
	for {
		nextFreeBlock := current.GetNext()
		if nextFreeBlock == nil || nextFreeBlock.Offset > offset {
			return current
		}
		current = nextFreeBlock
	}
}

func (p Page) getFreeBlockJustAfter(offset uint16) *freeBlock {
	current := p.getFirstFreeBlock()
	if current == nil {
		return nil
	} else if current.Offset > offset {
		return current
	}
	for {
		nextFreeBlock := current.GetNext()
		if nextFreeBlock == nil {
			return nil
		}
		if nextFreeBlock.Offset > offset {
			return nextFreeBlock
		}
		current = nextFreeBlock
	}
}

func (p Page) insertFreeBlockUnchecked(offset, blockSize, parentOffset uint16) {
	if parentOffset == InvalidOffset {
		rawHeader := p.GetRawHeader()
		firstFreeBlockOffset := rawHeader.GetFirstFreeBlockOffset()
		if firstFreeBlockOffset == InvalidOffset {
			rawHeader.SetFirstFreeBlockOffset(offset)
			fbh := newFreeBlockHeader()
			fbh.setNextFreeBlockOffset(InvalidOffset)
			fbh.setBlockSize(blockSize)
			copy(p[offset:offset+FREE_BLOCK_HEADER_SIZE], fbh)
		} else {
			if firstFreeBlockOffset < offset {
				p.insertFreeBlockUnchecked(offset, blockSize, firstFreeBlockOffset)
			}
			fbh := newFreeBlockHeader()
			fbh.setNextFreeBlockOffset(firstFreeBlockOffset)
			fbh.setBlockSize(blockSize)
			rawHeader.SetFirstFreeBlockOffset(offset)
			copy(p[offset:offset+FREE_BLOCK_HEADER_SIZE], fbh)
		}
	} else {
		parentFbh := p.getFreeBlockAtOffsetUnchecked(parentOffset)
		newNextOffset := parentFbh.getNextFreeBlockOffset()
		if newNextOffset < offset {
			p.insertFreeBlockUnchecked(offset, blockSize, newNextOffset)
		}
		parentFbh.setNextFreeBlockOffset(offset)
		fbh := newFreeBlockHeader()
		fbh.setNextFreeBlockOffset(newNextOffset)
		fbh.setBlockSize(blockSize)
		copy(p[offset:offset+FREE_BLOCK_HEADER_SIZE], fbh)
	}
}

func (p Page) getNthCellOffset(index uint16) (uint16, error) {
	if index >= p.GetRawHeader().GetCellsCount() {
		return 0, fmt.Errorf("cell index out of range")
	}
	return serializarion.DeserializeUint16(p[PAGE_HEADER_SIZE+2*index : PAGE_HEADER_SIZE+2*index+2]), nil
}

func (p Page) getNthCellOffsetUnchecked(index uint16) uint16 {
	return serializarion.DeserializeUint16(p[PAGE_HEADER_SIZE+2*index : PAGE_HEADER_SIZE+2*index+2])
}

func (p Page) isOffsetStartOfACell(offset uint16) bool {
	for i := uint16(0); i < p.GetRawHeader().GetCellsCount(); i++ {
		if p.getNthCellOffsetUnchecked(i) == offset {
			return true
		}
	}
	return false
}

func (p Page) getCellOffsetAt(index uint16) uint16 {
	offsetBytes := p[PAGE_HEADER_SIZE+2*index : PAGE_HEADER_SIZE+2*index+2]
	return serializarion.DeserializeUint16(offsetBytes)
}

func (p Page) GetCellAt(index uint16) (Cell, error) {
	cellOffset, err := p.getNthCellOffset(index)
	if err != nil {
		return nil, err
	}
	return p.getCellAtOffsetUnchecked(cellOffset), nil
}

func (p Page) GetCellAtOffset(cellOffset uint16) (Cell, error) {
	if !p.isOffsetStartOfACell(cellOffset) {
		return nil, fmt.Errorf("there is no cell at this offset")
	}
	payloadSize := serializarion.DeserializeUint16(p[cellOffset : cellOffset+2])
	cellPayload := p[cellOffset+2 : cellOffset+2+payloadSize]
	cell, _ := NewCell(cellPayload)
	return cell, nil
}

func (p Page) getCellAtOffsetUnchecked(cellOffset uint16) Cell {
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
		nthOffset := p.getNthCellOffsetUnchecked(i)
		tempCell := p.getCellAtOffsetUnchecked(nthOffset)
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

func (p Page) deleteCellOffsetFromList(cellLocation uint16) {
	cellsCount := p.GetRawHeader().GetCellsCount()
	for i := uint16(0); i < cellsCount; i++ {
		nthOffset := p.getNthCellOffsetUnchecked(i)
		if nthOffset == cellLocation {
			for j := i; j < cellsCount-1; j++ {
				copy(p[PAGE_HEADER_SIZE+2*j:PAGE_HEADER_SIZE+2*j+2], p[PAGE_HEADER_SIZE+2*j+2:PAGE_HEADER_SIZE+2*j+4])
			}
			break
		}
	}
}

func (p Page) GetCellsOffsets() []uint16 {
	cellsCount := p.GetRawHeader().GetCellsCount()
	cellOffsets := make([]uint16, cellsCount)
	for i := uint16(0); i < cellsCount; i++ {
		cellOffsets[i] = p.getNthCellOffsetUnchecked(i)
	}
	return cellOffsets
}

func (p Page) replaceCellOffsets(offsets []uint16) {
	offsetArrLength := uint16(len(offsets))
	for i := uint16(0); i < offsetArrLength; i++ {
		byted := serializarion.SerializeUint16(offsets[i])
		p[PAGE_HEADER_SIZE+2*i] = byted[0]
		p[PAGE_HEADER_SIZE+2*i+1] = byted[1]
	}
}

func (p Page) allocateBlockInEmptyAreas(size uint16) *freeBlock {
	//returns nil header if not able to allocate
	freeBlockObject := p.getFirstFreeBlock()

	for freeBlockObject != nil {
		if freeBlockObject.GetSize() >= size {
			return freeBlockObject
		}
		freeBlockObject = freeBlockObject.GetNext()
	}
	return nil
}

func (p Page) allocateNewBlock(size uint16) (offset uint16) {
	firstCellOffset := p.GetRawHeader().GetFirstCellOffset()
	if firstCellOffset == InvalidOffset {
		return PAGE_SIZE - size
	} else {
		return firstCellOffset - size
	}
}

//function to work with empty blocks
const FREE_BLOCK_HEADER_SIZE = 4

type freeBlockHeader []byte

type freeBlock struct {
	Offset       uint16
	header       freeBlockHeader
	ParentOffset uint16
	page         *Page
}

func (fb *freeBlock) GetBytes() []byte {
	return fb.header
}

func (fb *freeBlock) GetParentHeader() freeBlockHeader {
	if fb.ParentOffset == 0 {
		return nil
	}
	return []byte(*fb.page)[fb.ParentOffset : fb.ParentOffset+FREE_BLOCK_HEADER_SIZE]
}

func (fb *freeBlock) GetNext() *freeBlock {
	nextOffset := serializarion.DeserializeUint16(fb.header[2:4])
	return &freeBlock{
		Offset:       nextOffset,
		header:       []byte(*fb.page)[nextOffset : nextOffset+FREE_BLOCK_HEADER_SIZE],
		ParentOffset: fb.Offset,
		page:         fb.page,
	}
}

func (fb *freeBlock) GetSize() uint16 {
	return serializarion.DeserializeUint16(fb.header[0:2])
}

func (fb *freeBlock) SetSize(newSize uint16) {
	copy(fb.header[0:2], serializarion.SerializeUint16(newSize))
}

func newFreeBlockHeader() freeBlockHeader {
	newFreeBlockHeader := [4]byte{0, 0, 0, 0}
	return newFreeBlockHeader[:]
}

func (fbh freeBlockHeader) getBlockSize() uint16 { //including header!
	return serializarion.DeserializeUint16(fbh[0:2])
}

func (fbh freeBlockHeader) setNextFreeBlockOffset(value uint16) { //including header!
	bts := serializarion.SerializeUint16(value)
	fbh[2] = bts[0]
	fbh[3] = bts[1]
}
func (fbh freeBlockHeader) setBlockSize(value uint16) { //including header!
	bts := serializarion.SerializeUint16(value)
	fbh[0] = bts[0]
	fbh[1] = bts[1]
}

func (fbh freeBlockHeader) getNextFreeBlockOffset() uint16 { //including header!
	return serializarion.DeserializeUint16(fbh[2:4])
}

func (p Page) getFirstFreeBlockHeader() freeBlockHeader {
	firstFreeBlockOffset := p.GetRawHeader().GetFirstFreeBlockOffset()
	if firstFreeBlockOffset == 0 {
		return nil
	}
	return p.getFreeBlockAtOffsetUnchecked(firstFreeBlockOffset)
}

func (p Page) getFirstFreeBlock() *freeBlock {
	firstFreeBlockOffset := p.GetRawHeader().GetFirstFreeBlockOffset()
	if firstFreeBlockOffset == InvalidOffset {
		return nil
	}
	headerSlice := []byte(p[firstFreeBlockOffset : firstFreeBlockOffset+4])
	return &freeBlock{
		Offset:       firstFreeBlockOffset,
		header:       headerSlice,
		ParentOffset: 0,
		page:         &p,
	}
}

func (p Page) getFreeBlockAtOffsetUnchecked(offset uint16) freeBlockHeader {
	return freeBlockHeader(p[offset : offset+4])
}

func (p Page) putCellInFreeBlockUnchecked(cell Cell, freeBlockObject *freeBlock) uint16 {
	offset := freeBlockObject.Offset
	freeBlockEndOffset := freeBlockObject.Offset + freeBlockObject.GetSize()
	cellSize := cell.GetSize()

	p.shrinkFreeBlock(freeBlockObject, cellSize)
	copy(p[freeBlockEndOffset-cellSize:freeBlockEndOffset], cell)

	return offset
}

func (p Page) getFreeBlockOffset(freeBlockParent freeBlockHeader) uint16 {
	var offset uint16
	if freeBlockParent == nil {
		offset = p.GetRawHeader().GetFirstFreeBlockOffset()
	} else {
		offset = freeBlockParent.getNextFreeBlockOffset()
	}
	return offset
}

func (p Page) shrinkFreeBlock(freeBlockObject *freeBlock, shrink uint16) {
	newSize := freeBlockObject.GetSize() - shrink
	if newSize > FREE_BLOCK_HEADER_SIZE {
		freeBlockObject.SetSize(newSize)
	} else {
		p.GetRawHeader().IncreaseFragmentedBytesCount(uint8(newSize))
		if freeBlockObject.ParentOffset != InvalidOffset {
			freeBlockObject.GetParentHeader().setNextFreeBlockOffset(freeBlockObject.header.getNextFreeBlockOffset())
		} else {
			p.GetRawHeader().SetFirstFreeBlockOffset(InvalidOffset)
		}
	}
}

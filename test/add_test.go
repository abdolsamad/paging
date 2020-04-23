package test

import (
	"testing"

	"github.com/abdolsamad/paging/paging"
)

func comparator(cell1 paging.Cell, cell2 paging.Cell) bool {
	return cell1.GetPayload()[0] < cell1.GetPayload()[1]
}

func TestOffset(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	offset, _ := p.AddCell(cell0, comparator)

	expected := paging.PAGE_SIZE - cell0.GetSize()
	got := offset
	if got != expected {
		t.Errorf("Expected an offset of %d but got %d instead", expected, got)
	}

	offset, _ = p.AddCell(cell1, comparator)

	expected = paging.PAGE_SIZE - cell0.GetSize() - cell1.GetSize()
	got = offset
	if got != expected {
		t.Errorf("Expected an offset of %d but got %d instead", expected, got)
	}
}

func TestFirstCellsOffset(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})
	p.AddCell(cell0, comparator)
	p.AddCell(cell1, comparator)
	p.AddCell(cell2, comparator)
	expected := paging.PAGE_SIZE - cell0.GetSize() - cell1.GetSize() - cell2.GetSize()
	got := p.GetRawHeader().GetFirstCellOffset()
	if got != expected {
		t.Errorf("Expected first cell offset of %d but got %d instead", expected, got)
	}
}

func TestAddUpdateCellsCount(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})

	p.AddCell(cell0, comparator)
	expected := uint16(1)
	got := p.GetRawHeader().GetCellsCount()
	if got != expected {
		t.Errorf("Expected cell count of %d but got %d instead", expected, got)
	}

	p.AddCell(cell1, comparator)
	expected = uint16(2)
	got = p.GetRawHeader().GetCellsCount()
	if got != expected {
		t.Errorf("Expected cell count of %d but got %d instead", expected, got)
	}

	p.AddCell(cell2, comparator)
	expected = uint16(3)
	got = p.GetRawHeader().GetCellsCount()
	if got != expected {
		t.Errorf("Expected cell count of %d but got %d instead", expected, got)
	}
}

func TestAddUpdateFragmentsCount(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})
	p.AddCell(cell0, comparator)
	p.AddCell(cell1, comparator)
	p.AddCell(cell2, comparator)
	expected := uint8(0)
	got := p.GetRawHeader().GetFragmentedBytesCount()
	if got != expected {
		t.Errorf("Expected fragmented bytes count of %d but got %d instead", expected, got)
	}
}

func TestFreeHeadersOffset(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})
	p.AddCell(cell0, comparator)
	p.AddCell(cell1, comparator)
	p.AddCell(cell2, comparator)
	expected := paging.InvalidOffset
	got := p.GetRawHeader().GetFirstFreeBlockOffset()
	if got != expected {
		t.Errorf("Expected first free block offset of %d but got %d instead", expected, got)
	}
}

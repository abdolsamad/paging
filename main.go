package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"os"
	"pager/paiging"
	"pager/serializarion/gps_ad_id"
	"pager/serializarion/protos"
	"time"
)

type PageFlag byte

const (
	PAGE_FLAG_LEAF = PageFlag(8)
)

type PageHeader struct {
	Flags                PageFlag
	FirstFreeBlockOffset uint16
	CellsCount           uint16
	FirstCellOffset      uint16
	FragmentedBytesCount uint8
}

type Page struct {
	header PageHeader
}

func writeZeroToFile(file *os.File, bytesCount uint32) {
	zeroByte := [1024]byte{}
	now := time.Now()
	for i := uint32(0); i < bytesCount; i++ {
		file.Write(zeroByte[:])
	}
	err := file.Sync()
	if err != nil {
		fmt.Println("Error in syncing file!")
		return
	}
	fmt.Printf("Write completed in %d ms\n", time.Since(now).Milliseconds())
}
func seekTest() {
	path := "d:\\large_file2.bin"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeExclusive)
	if err != nil {
		fmt.Printf("Error opening file(error: %s)\n", err)
		return
	}
	writeZeroToFile(file, 1024)
	printFileSize(file)
	now := time.Now()
	n, err := file.WriteAt([]byte{1, 1, 1, 1}, 0)
	if err != nil {
		fmt.Printf("Error updating file(error: %s)\n", err)
		return
	} else {
		err = file.Sync()
		fmt.Printf("%d bytes updated\n", n)
		printFileSize(file)
	}
	if err != nil {
		fmt.Printf("Error syncing file(error: %s)\n", err)
		return
	}
	fmt.Printf("File updated in %d ms\n", time.Since(now).Milliseconds())
	file.Close()
}

const ChunkSize = 100 * 1024 * 1024

type LargeType [ChunkSize]byte

func ignore(b byte) {
	fmt.Println(b)
}
func UserDataToCell(userData *protos.UserData) (paiging.Cell, error) {
	bytesU1, err := proto.Marshal(userData)
	if err != nil {
		return nil, err
	}
	cell, err := paiging.NewCell(bytesU1)
	if err != nil {
		return nil, err
	}
	return cell, nil
}
func CellToUserData(cell paiging.Cell) (*protos.UserData, error) {
	ud := &protos.UserData{}
	err := proto.Unmarshal(cell[2:], ud)
	if err != nil {
		return nil, err
	}
	return ud, nil
}
func AddUserData(page paiging.Page, userData *protos.UserData) error {
	cell, err := UserDataToCell(userData)
	if err != nil {
		return err
	}

	_, err = page.AddCell(cell, paiging.GpsAdIdComparator)
	if err != nil {
		return err
	}
	return nil
}
func main() {
	start := time.Now()
	page := paiging.NewPage()
	gpsAdId1, err := gps_ad_id.NewFromString("3ee39630-8b64-4510-8780-ff25514a1188")
	if err != nil {
		fmt.Println(err)
		return
	}
	u1 := &protos.UserData{
		GpsAdId:      gpsAdId1,
		CurrentApps:  []uint32{10, 12, 14, 16},
		PreviousApps: []uint32{156861, 49849, 135168},
	}
	err = AddUserData(page, u1)
	if err != nil {
		fmt.Println(err)
		return
	}
	gpsAdId2, err := gps_ad_id.NewFromString("3fe39630-8b64-4510-8780-ff25514a1188")
	if err != nil {
		fmt.Println(err)
		return
	}
	u2 := &protos.UserData{
		GpsAdId:      gpsAdId2,
		CurrentApps:  []uint32{18, 56, 122, 36},
		PreviousApps: []uint32{999885, 6688, 64456, 1335133, 13813},
	}
	err = AddUserData(page, u2)
	if err != nil {
		fmt.Println(err)
		return
	}
	cell1 := page.GetCellAt(1)
	userData, err := CellToUserData(cell1)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(page)
	fmt.Printf("The whole thing took: %d us\n", time.Since(start).Microseconds())
	PrintUserData(userData)
}
func PrintUserData(userData *protos.UserData) {
	gpsAdId := userData.GetGpsAdId()
	gpsAdIdString, _ := gps_ad_id.GpsAdId(gpsAdId).ToString()
	fmt.Printf("GpsAdId: %s\n", gpsAdIdString)
	currentApps := userData.GetCurrentApps()
	fmt.Print("Current apps: ")
	fmt.Println(currentApps)
	previousApps := userData.GetPreviousApps()
	fmt.Print("Previous apps: ")
	fmt.Println(previousApps)
}
func testWrite1Mb(file *os.File) {
	chunk := make([]byte, 1024)
	_, err := file.Write(chunk)
	if err != nil {
		fmt.Printf("Error in appending: %s", err)
	}
	err = file.Sync()
	if err != nil {
		fmt.Printf("Error in syncing: %s", err)
	}
}

func printFileSize(file *os.File) {

	stats, _ := file.Stat()
	fmt.Printf("file size is: %d\n", stats.Size())
}

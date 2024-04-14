package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Designed by Joseph Craig, 2024
// Define structures for MBR and GPT

type Partition struct {
	BootFlag     uint8
	StartCHS     [3]uint8
	Type         uint8
	EndCHS       [3]uint8
	StartLBA     uint32
	TotalSectors uint32
}

type MBR struct {
	BootstrapCode [446]byte
	Partitions    [4]Partition
	Signature     uint16
}

type GPTHeader struct {
	Signature           [8]byte
	Revision            uint32
	HeaderSize          uint32
	CRC32               uint32
	Reserved            uint32
	CurrentLBA          uint64
	BackupLBA           uint64
	FirstUsableLBA      uint64
	LastUsableLBA       uint64
	DiskGUID            [16]byte
	PartitionEntryLBA   uint64
	NumberOfPartitions  uint32
	PartitionEntrySize  uint32
	PartitionEntryCRC32 uint32
}

func processMBR(file *os.File) {
	// Seek to the beginning of the file
	_, err := file.Seek(0, 0)
	if err != nil {
		fmt.Printf("Error seeking file: %v\n", err)
		return
	}

	mbr := MBR{}
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Printf("Error reading MBR: %v\n", err)
		return
	}

	fmt.Println("MBR Partition Table:")
	for i, p := range mbr.Partitions {
		fmt.Printf("Partition %d:\n", i+1)
		fmt.Printf("\tBoot Flag: 0x%X\n", p.BootFlag)
		fmt.Printf("\tType: 0x%X\n", p.Type)
		fmt.Printf("\tStart LBA: %d\n", p.StartLBA)
		fmt.Printf("\tTotal Sectors: %d\n", p.TotalSectors)
	}
}

func processGPT(file *os.File) {
	// Seek to LBA 1 where the GPT header is located
	_, err := file.Seek(512, 0)
	if err != nil {
		fmt.Printf("Error seeking to GPT header: %v\n", err)
		return
	}

	gptHeader := GPTHeader{}
	err = binary.Read(file, binary.LittleEndian, &gptHeader)
	if err != nil {
		fmt.Printf("Error reading GPT header: %v\n", err)
		return
	}

	fmt.Println("GPT Header Information:")
	fmt.Printf("\tDisk GUID: %x\n", gptHeader.DiskGUID)
	fmt.Printf("\tPartition Entry LBA: %d\n", gptHeader.PartitionEntryLBA)
	fmt.Printf("\tNumber Of Partitions: %d\n", gptHeader.NumberOfPartitions)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <image file>")
		os.Exit(1)
	}

	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read first 512 bytes to check for MBR signature or protective MBR for GPT
	mbr := MBR{}
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Printf("Error reading initial sector: %v\n", err)
		os.Exit(1)
	}

	if mbr.Signature == 0xAA55 {
		// Check if it's a protective MBR for GPT
		if mbr.Partitions[0].Type == 0xEE {
			fmt.Println("Detected GPT Disk. Processing GPT...")
			processGPT(file)
		} else {
			fmt.Println("Detected MBR Disk. Processing MBR...")
			processMBR(file)
		}
	} else {
		fmt.Println("Unknown disk format or missing valid boot signature.")
	}
}

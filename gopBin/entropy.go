package gopBin

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
)

type Entropy interface {
	ParseBytes(string) error
	ComputeEntropy() float64
	PrintEntropy()
}

type File struct {
	Name    string
	Content []byte
}

type SimpleFile struct {
	File
}

type PEFile struct {
	File
	Sections map[string][]byte
}

// Entry point
func GetFileEntropyHandle(filename string) (Entropy, error) {
	var f Entropy

	fOpen, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fOpen.Close()

	MagicByte := make([]byte, 3)
	_, err = fOpen.ReadAt(MagicByte, 0)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(MagicByte[0:2], []byte{0x4D, 0x5A}) {
		f = &PEFile{
			File:     File{},
			Sections: make(map[string][]byte),
		}
	} else {
		f = &SimpleFile{}
	}

	return f, nil
}

func GetEntropy(data []byte) float64 {
	set := make(map[byte]int)
	for i := 0; i < 256; i++ {
		set[byte(i)] = 0
	}

	for _, i := range data {
		set[i]++
	}

	entropy := float64(0)

	for _, x := range set {
		if x == 0 {
			continue
		}

		p := float64(x) / float64(len(data))

		entropy -= p * math.Log2(p)
	}

	return entropy
}

func (f *File) ParseBytes(filename string) error {
	f.Name = filename

	fOpen, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fOpen.Close()

	f.Content, err = io.ReadAll(fOpen)
	if err != nil {
		return err
	}

	return nil
}

func (f *File) ComputeEntropy() float64 {
	return GetEntropy(f.Content)
}

func (f *File) PrintEntropy() {
	fmt.Printf("Entropy of %s : %f\n", f.Name, f.ComputeEntropy())
}

// func (f *SimpleFile) ComputeEntropy() float64 {
// 	return f.File.ComputeEntropy()
// }

// func (f *PEFile) ParseBytes(filename string) error {
// 	f.File.ParseBytes(filename)

// 	peFile, err := pe.Open(filename)
// 	if err != nil {
// 		return err
// 	}
// 	defer peFile.Close()

// 	for _, section := range peFile.Sections {

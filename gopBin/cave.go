package gopbin

import (
	"debug/pe"
	"fmt"
)

func GetPECaves(filename string, minCaveSize int) error {
	peFile, err := pe.Open(filename)
	if err != nil {
		return err
	}
	defer peFile.Close()

	// For each sections, list the caves by starting offsets and get its length
	caves := make(map[*pe.Section]map[uint32]int)

	for _, section := range peFile.Sections {
		data, err := section.Data()
		if err != nil {
			fmt.Printf("[!] Error trying read data of section %s\n", section.Name)
		}

		// We can read data, initialise the map
		caves[section] = make(map[uint32]int)

		length := 0
		offset := uint32(0)

		for position, value := range data {
			if length == 0 {
				if value == 0x00 {
					offset = uint32(position)
					length++
				}
			} else if length > 0 {
				if value == 0x00 {
					length++
				} else {
					if length >= minCaveSize {
						caves[section][offset] = length
					}
					length = 0
					offset = 0
				}
			}
			continue
		}
	}

	for section, cave := range caves {
		fmt.Printf("%s :\n", section.Name)
		for offset, length := range cave {
			fmt.Printf("\tfile offset : 0x%08x (%d) - section offset %08x (%d) - size %d bytes\n", section.Offset+uint32(offset), section.Offset+uint32(offset), offset, offset, length)
		}
	}

	return nil
}

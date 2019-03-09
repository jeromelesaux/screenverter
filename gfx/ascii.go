package gfx

import (
	"encoding/binary"
	"fmt"
	"github.com/jeromelesaux/m4client/cpc"
	"image/color"
	"os"
	"path/filepath"
	"strings"
)

var ByteToken = "BYTE"

func Ascii(filePath, dirPath string, data []byte, p color.Palette, noAmsdosHeader, isCpcPlus bool) error {
	fmt.Fprintf(os.Stdout, "Writing ascii file (%s) data length (%d)\n", filePath, len(data))
	var out string
	var i int
	filename := filepath.Base(filePath)
	extension := filepath.Ext(filename)
	cpcFilename := strings.ToUpper(strings.Replace(filename, extension, ".TXT", -1))
	out += "# Screen " + cpcFilename + "\n.screen:\n"
	for i = 0; i < len(data); i += 8 {
		out += fmt.Sprintf("%s #%0.2x, #%0.2x, #%0.2x, #%0.2x, #%0.2x, #%0.2x, #%0.2x, #%0.2x\n", ByteToken,
			data[i], data[i+1], data[i+2], data[i+3],
			data[i+4], data[i+5], data[i+6], data[i+7])
	}
	out += "# Palette " + cpcFilename + "\n.palette:\n" + ByteToken + " "

	if isCpcPlus {
		for i := 0; i < len(p); i++ {
			cp := NewCpcPlusColor(p[i])
			v := cp.Value()
			out += fmt.Sprintf("#%.2x, #%.2x", byte(v), byte(v>>8))
			if (i+1)%8 == 0 && i+1 < len(p) {
				out += "\n" + ByteToken + " "
			} else {
				if i+1 < len(p) {
					out += ", "
				}
			}
		}
	} else {
		for i := 0; i < len(p); i++ {
			v, err := HardwareValues(p[i])
			if err == nil {
				out += fmt.Sprintf("#%0.2x", v[0])
				if (i+1)%8 == 0 && i+1 < len(p) {
					out += "\n" + ByteToken + " "
				} else {
					if i+1 < len(p) {
						out += ", "
					}
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error while getting the hardware values for color %v, error :%d\n", p[0], err)
			}
		}
	}
	//fmt.Fprintf(os.Stdout,"%s",out)
	header := cpc.CpcHead{Type: 0, User: 0, Address: 0x0, Exec: 0x0,
		Size:        uint16(len(out)),
		Size2:       uint16(len(out)),
		LogicalSize: uint16(len(out))}

	copy(header.Filename[:], strings.Replace(cpcFilename, ".", "", -1))
	header.Checksum = uint16(header.ComputedChecksum16())
	fmt.Fprintf(os.Stderr, "Header length %d\n", binary.Size(header))
	fw, err := os.Create(dirPath + string(filepath.Separator) + cpcFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating file (%s) error :%s\n", cpcFilename, err)
		return err
	}
	if !noAmsdosHeader {
		binary.Write(fw, binary.LittleEndian, header)
	}
	binary.Write(fw, binary.LittleEndian, []byte(out))
	fw.Close()
	return nil
}
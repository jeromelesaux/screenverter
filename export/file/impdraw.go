package file

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/jeromelesaux/m4client/cpc"
	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/export"
)

/*
 initialisation du raster à faire qu'un fois
 ld bc,#7F00
 ld hl,#8E8D
 ld d,#8c
 out (c), c
 out (c), d // out du mode 0
 code assembleur pour déclencher un split raster
 ld c,pen: out (c), c // selection du pen à modifier
 ld a,#54: out (c), a // on envoie la couleur que l'on veut
 ld a,#54: out (c), a // valeur hardware
 ld a,#54: out (c), a
 ld a,#54: out (c), a
 ld a,#54: out (c), a
 ld a,#54: out (c), a
 ld a,#54: out (c), a
 ld a,#54: out (c), a
 ld a,(bc) // restitution du mode
 out (c),a // resitution de la couleur depuis la palette en #7F00
 out (c),h // restitution du mode

 // 256 splits rasters maximum
*/

/*
 initialisation du raster à faire qu'un fois
 ld bc,#7F00
 ld hl,#8E8D
 ld d,#8c
 out (c), c
 out (c), d // out du mode 0
*/
var splitRasterSetMode = []byte{0x01, 0x00, 0x7F, 0x21, 0x8D, 0x8E, 0x16, 0x8C, 0xED, 0x49, 0xED, 0x69}

/*
ld c,pen: out (c), c // selection du pen à modifier
*/
var splitRasterSelectPen = []byte{0x0E, 0x00, 0xED, 0x49}

/*
ld a,#54: out (c), a
*/
var splitRasterSetColor = []byte{0x3E, 0x54, 0xED, 0x79}

/*
 ld a,(bc) // restitution du mode
 out (c),a // resitution de la couleur depuis la palette en #7F00
 out (c),h // restitution du mode
*/
var splitRasterRestore = []byte{0x0A, 0xED, 0x79, 0xED, 0x61}

func ExportSplitRaster(filename string, p color.Palette, rasters *constants.SplitRasterScreen, exportType *export.ExportType) error {

	output := make([]byte, 0)
	// set the init split raster routine assembled opcode
	output = append(output, splitRasterSetMode...)
	fmt.Fprintf(os.Stdout, "{%d} splits rasters found\n", len(rasters.Values))
	for _, v := range rasters.Values {
		// set the set mode assembled opcode
		setPal := splitRasterSetMode
		setPal[1] = byte(v.PaletteIndex[0])
		//fmt.Fprintf(os.Stdout, "Set pen {%d}\n", v.PaletteIndex[0])
		output = append(output, setPal...)
		// set the set color assembled opcode
		for i, h := range v.HardwareColor {
			if i%2 == 0 {
				setColor := splitRasterSetColor
				setColor[1] = byte(h)
				//fmt.Fprintf(os.Stdout, "Set color {%d}\n", h)
				output = append(output, setColor...)
			}
		}
	}
	output = append(output, splitRasterRestore...)

	header := cpc.CpcHead{Type: 0, User: 0, Address: 0x170, Exec: 0x0,
		Size:        uint16(binary.Size(output)),
		Size2:       uint16(binary.Size(output)),
		LogicalSize: uint16(binary.Size(output))}
	copy(header.Filename[:], exportType.GetAmsdosFilename(filename, ".SPL"))
	header.Checksum = uint16(header.ComputedChecksum16())
	basicPath := filepath.Join(exportType.OutputPath, exportType.GetAmsdosFilename(filename, ".SPL"))
	fw, err := os.Create(basicPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating file (%s) error :%s\n", basicPath, err)
		return err
	}
	if !exportType.NoAmsdosHeader {
		binary.Write(fw, binary.LittleEndian, header)
	}
	binary.Write(fw, binary.LittleEndian, output)
	fw.Close()

	exportType.AddFile(basicPath)
	return nil
}

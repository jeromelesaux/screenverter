package animate

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/convert"
	"github.com/jeromelesaux/martine/export"
	"github.com/jeromelesaux/martine/export/file"
	"github.com/jeromelesaux/martine/gfx/common"
	"github.com/jeromelesaux/martine/gfx/transformation"
	zx0 "github.com/jeromelesaux/zx0/encode"
)

func DeltaMotif(gitFilepath string, ex *export.ExportType, threshold int, initialAddress uint16, mode uint8) error {

	fr, err := os.Open(gitFilepath)
	if err != nil {
		return err
	}
	defer fr.Close()
	gifImages, err := gif.DecodeAll(fr)
	if err != nil {
		return err
	}
	images := ConvertToImage(*gifImages)
	// convertion de la palette en mode 1
	size := constants.Mode1

	screens := make([]*image.NRGBA, 0)
	// resize all images
	for _, v := range images {
		out := convert.Resize(v, size, imaging.NearestNeighbor)
		screens = append(screens, out)
	}

	// downgrading palette
	customPalette, _, err := convert.DowngradingPalette(screens[0], size, ex.CpcPlus)
	if err != nil {
		return err
	}
	// converting all screens
	for index, v := range screens {
		_, out := convert.DowngradingWithPalette(v, customPalette)
		screens[index] = out
	}

	// recuperation des motifs
	a := transformation.AnalyzeTilesBoard(screens[0], constants.Size{Width: 4, Height: 4})
	refBoard := a.ReduceTilesNumber(float64(threshold))
	btc := make([][]transformation.BoardTile, 0)
	btc = append(btc, refBoard)
	refTiles := transformation.GetUniqTiles(refBoard)
	a.Image("../../test/motifs/first.png", refBoard, a.ImageSize)
	// application des motifs sur toutes les images
	for i := 1; i < len(screens); i++ {
		ab := transformation.AnalyzeTilesBoardWithTiles(screens[i], constants.Size{Width: 4, Height: 4}, refTiles)
		board := ab.BoardTiles
		btc = append(btc, board)
		ab.Image(fmt.Sprintf("../../test/motifs/%.2d.png", i), board, a.ImageSize)
	}

	motifs := make([][]byte, 0)

	/* conversion des sprites en mode cpc */
	for i := 0; i < len(refTiles); i++ {
		sprt := (&refTiles[i]).Image()
		data, _, _, _ := common.ToSprite(sprt, customPalette, refTiles[i].Size, 1, ex)
		motifs = append(motifs, data)
	}

	/* calcul des coordonnées */
	deltas := make([][]byte, 0)
	for _, v := range btc {
		//nbCoords := size.Width * size.Height / 2 / 4
		delta := make([]byte, 0)
		index := 0
		pixel := 0
		for j := 0; j < size.Height; j += 4 {
			for i := 0; i < size.Width; i += 4 {
				var pos int
				if t := transformation.GetTile(v, i, j); t != nil {
					pos = transformation.GetTilePostion(t, refTiles)

				}
				if index == 2 {
					delta = append(delta, byte(pixel))
					pixel = 0
					index = 0
				}
				if index == 0 {
					pixel += (pos << 4)
				} else {
					pixel += pos
				}
				index++
			}
		}
		deltas = append(deltas, delta)
	}
	filename := string(ex.OsFilename(".asm"))
	return exportDeltaMotif(deltas, motifs, customPalette, ex, ex.OutputPath+string(filepath.Separator)+filename)
}

func exportDeltaMotif(images [][]byte, motifs [][]byte, p color.Palette, ex *export.ExportType, filename string) error {
	var deltaCode string
	for i := 0; i < len(images); i++ {
		deltaCode += fmt.Sprintf("delta%.2d\n", i)
		encoded := zx0.Encode(images[i])
		deltaCode += file.FormatAssemblyDatabyte(encoded, "\n")
	}

	for i := 0; i < len(motifs); i++ {
		deltaCode += fmt.Sprintf("motif%.2d\n", i)
		deltaCode += file.FormatAssemblyDatabyte(motifs[i], "\n")
	}

	deltaCode += "palette:\n" + file.ByteToken + " "
	deltaCode += file.FormatAssemblyBasicPalette(p, "\n")

	//	fmt.Printf("%s", deltaCode)
	fw, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fw.Close()
	fw.WriteString(deltaCode)
	return nil
}

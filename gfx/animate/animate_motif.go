package animate

import (
	"fmt"
	"image"
	"image/gif"
	"os"

	"github.com/disintegration/imaging"
	"github.com/jeromelesaux/martine/constants"
	"github.com/jeromelesaux/martine/convert"
	"github.com/jeromelesaux/martine/export"
	"github.com/jeromelesaux/martine/gfx/transformation"
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

	return nil
}

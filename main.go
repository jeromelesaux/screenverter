package main

import (
	"flag"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/jeromelesaux/martine/convert"
	"github.com/jeromelesaux/martine/gfx"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	byteStatement   = flag.String("s", "", "Byte statement to replace in ascii export (default is BYTE), you can replace or instance by defb")
	picturePath     = flag.String("i", "", "Picture path of the input file.")
	width           = flag.Int("w", -1, "Custom output width in pixels.")
	height          = flag.Int("h", -1, "Custom output height in pixels.")
	mode            = flag.Int("m", -1, "Output mode to use :\n\t0 for mode0\n\t1 for mode1\n\t2 for mode2\n\tand add -f option for overscan export.\n\t")
	output          = flag.String("o", "", "Output directory")
	overscan        = flag.Bool("f", false, "Overscan mode (default no overscan)")
	resizeAlgorithm = flag.Int("a", 1, "Algorithm to resize the image (available : \n\t1: NearestNeighbor (default)\n\t2: CatmullRom\n\t3: Lanczos\n\t4: Linear\n\t5: Box\n\t6: Hermite\n\t7: BSpline\n\t8: Hamming\n\t9: Hann\n\t10: Gaussian\n\t11: Blackman\n\t12: Bartlett\n\t13: Welch\n\t14: Cosine\n\t")
	help            = flag.Bool("help", false, "Display help message")
	noAmsdosHeader  = flag.Bool("n", false, "no amsdos header for all files (default amsdos header added).")
	plusMode        = flag.Bool("p", false, "Plus mode (means generate an image for CPC Plus Screen)")
	rollMode        = flag.Bool("roll", false, "Roll mode allow to walk and walk into the input file.")
	iterations      = flag.Int("iter", -1, "Iterations number to walk in tile mode")
	rra             = flag.Int("rra", -1, "bit rotation on the right and keep pixels")
	rla             = flag.Int("rla", -1, "bit rotation on the left and keep pixels")
	sra             = flag.Int("sra", -1, "bit rotation on the right and lost pixels")
	sla             = flag.Int("sla", -1, "bit rotation on the left and lost pixels")
	losthigh        = flag.Int("losthigh", -1, "bit rotation on the top and lost pixels")
	lostlow         = flag.Int("lostlow", -1, "bit rotation on the bottom and lost pixels")
	keephigh        = flag.Int("keephigh", -1, "bit rotation on the top and keep pixels")
	keeplow         = flag.Int("keeplow", -1, "bit rotation on the bottom and keep pixels")
	version         = "0.5"
)

func usage() {
	fmt.Fprintf(os.Stdout, "martine convert (jpeg, png format) image to Amstrad cpc screen (even overscan)\n")
	fmt.Fprintf(os.Stdout, "By Impact Sid (Version:%s)\n", version)
	fmt.Fprintf(os.Stdout, "Special thanks to @Ast (for his support), @Siko and @Tronic for ideas\n")
	fmt.Fprintf(os.Stdout, "usage :\n\n")
	flag.PrintDefaults()
	os.Exit(-1)
}

func main() {
	var size gfx.Size
	var filename, extension string
	var customDimension bool
	var screenMode uint8
	flag.Parse()

	if *help {
		usage()
	}

	// picture path to convert
	if *picturePath == "" {
		usage()
	}
	filename = filepath.Base(*picturePath)
	extension = filepath.Ext(*picturePath)

	// output directory to store results
	if *output != "" {
		fi, err := os.Stat(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting directory informations :%v, Quiting\n", err)
			os.Exit(-2)
		}

		if !fi.IsDir() {
			fmt.Fprintf(os.Stderr, "%s is not a directory will store in current directory\n", *output)
			*output = "./"
		}
	} else {
		*output = "./"
	}

	if *mode == -1 {
		fmt.Fprintf(os.Stderr, "No output mode defined can not choose. Quiting\n")
		usage()
	}
	switch *mode {
	case 0:
		size = gfx.Mode0
		screenMode = 0
		if *overscan {
			size = gfx.OverscanMode0
		}
	case 1:
		size = gfx.Mode1
		screenMode = 1
		if *overscan {
			size = gfx.OverscanMode1
		}
	case 2:
		screenMode = 2
		size = gfx.Mode2
		if *overscan {
			size = gfx.OverscanMode2
		}
	default:
		if *height == -1 && *width == -1 {
			fmt.Fprintf(os.Stderr, "mode %d not defined and no custom width or height\n", *mode)
			usage()
		}
	}
	if *height != -1 {
		customDimension = true
		size.Height = *height
		if *width != -1 {
			size.Width = *width
		} else {
			size.Width = 0
		}
	}
	if *width != -1 {
		customDimension = true
		size.Width = *width
		if *height != -1 {
			size.Height = *height
		} else {
			size.Height = 0
		}
	}

	if *byteStatement != "" {
		gfx.ByteToken = *byteStatement
	}

	fmt.Fprintf(os.Stdout, "Informations :\n%s", size.ToString())

	f, err := os.Open(*picturePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening file %s, error %v\n", *picturePath, err)
		os.Exit(-2)
	}
	defer f.Close()
	in, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot decode the image %s error %v", *picturePath, err)
		os.Exit(-2)
	}

	fmt.Fprintf(os.Stderr, "Filename :%s, extension:%s\n", filename, extension)

	var resizeAlgo imaging.ResampleFilter
	switch *resizeAlgorithm {
	case 1:
		resizeAlgo = imaging.NearestNeighbor
	case 2:
		resizeAlgo = imaging.CatmullRom
	case 3:
		resizeAlgo = imaging.Lanczos
	case 4:
		resizeAlgo = imaging.Linear
	case 5:
		resizeAlgo = imaging.Box
	case 6:
		resizeAlgo = imaging.Hermite
	case 7:
		resizeAlgo = imaging.BSpline
	case 8:
		resizeAlgo = imaging.Hamming
	case 9:
		resizeAlgo = imaging.Hann
	case 10:
		resizeAlgo = imaging.Gaussian
	case 11:
		resizeAlgo = imaging.Blackman
	case 12:
		resizeAlgo = imaging.Bartlett
	case 13:
		resizeAlgo = imaging.Welch
	case 14:
		resizeAlgo = imaging.Cosine
	default:
		resizeAlgo = imaging.NearestNeighbor
	}

	out := convert.Resize(in, size, resizeAlgo)
	fmt.Fprintf(os.Stdout, "Saving resized image into (%s)\n", filename+"_resized.png")
	if err := gfx.Png(*output+string(filepath.Separator)+filename+"_resized.png", out); err != nil {
		os.Exit(-2)
	}

	newPalette, downgraded, err := convert.DowngradingPalette(out, size, *plusMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot downgrade colors palette for this image %s\n", *picturePath)
	}

	fmt.Fprintf(os.Stdout, "Saving downgraded image into (%s)\n", filename+"_down.png")
	if err := gfx.Png(*output+string(filepath.Separator)+filename+"_down.png", downgraded); err != nil {
		os.Exit(-2)
	}

	if *rollMode {
		// create downgraded palette image with rra pixels rotated
		// and call n iterations spritetransform with this input generated image
		// save the rotated image as png
		if *rla != -1 || *sla != -1 {
			fmt.Fprintf(os.Stdout, "RLA/SLA: Iterations (%d)\n", *iterations)
			for i := 0; i < *iterations; i++ {
				nbPixels := 0
				if *rla != -1 {
					nbPixels = (*rla * (1 + i))
				} else {
					if *sla != -1 {
						nbPixels = (*sla * (1 + i))
					}
				}
				im := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{downgraded.Bounds().Max.X, downgraded.Bounds().Max.Y}})
				y2 := 0
				for y := downgraded.Bounds().Min.Y; y < downgraded.Bounds().Max.Y; y++ {
					x2 := 0
					for x := downgraded.Bounds().Min.X + nbPixels; x < downgraded.Bounds().Max.X; x++ {
						im.Set(x2, y2, downgraded.At(x, y))
						x2++
					}
					y2++
				}
				if *rla != -1 {
					y2 = 0
					for y := downgraded.Bounds().Min.Y; y < downgraded.Bounds().Max.Y; y++ {
						x2 := downgraded.Bounds().Max.X - nbPixels
						for x := downgraded.Bounds().Min.X; x < nbPixels; x++ {
							im.Set(x2, y2, downgraded.At(x, y))
							x2++
						}
						y2++
					}
				}
				newFilename := strconv.Itoa(i) + strings.TrimSuffix(filename, path.Ext(filename)) + ".png"
				fmt.Fprintf(os.Stdout, "Saving downgraded image iteration (%d) into (%s)\n", i, newFilename)
				gfx.Png(*output+string(filepath.Separator)+newFilename, im)
				fmt.Fprintf(os.Stdout, "Tranform image in sprite iteration (%d)\n", i)
				gfx.SpriteTransform(im, newPalette, size, screenMode, newFilename, *output, *noAmsdosHeader, *plusMode)
			}
		} else {
			if *rra != -1 || *sra != -1 {
				fmt.Fprintf(os.Stdout, "RRA/SRA: Iterations (%d)\n", *iterations)

				for i := 0; i < *iterations; i++ {
					nbPixels := 0
					if *rra != -1 {
						nbPixels = (*rra * (1 + i))
					} else {
						if *sra != -1 {
							nbPixels = (*sra * (1 + i))
						}
					}
					im := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{downgraded.Bounds().Max.X, downgraded.Bounds().Max.Y}})
					y2 := 0
					for y := downgraded.Bounds().Min.Y; y < downgraded.Bounds().Max.Y; y++ {
						x2 := nbPixels
						for x := downgraded.Bounds().Min.X; x < downgraded.Bounds().Max.X-nbPixels; x++ {
							im.Set(x2, y2, downgraded.At(x, y))
							x2++
						}
						y2++
					}
					if *rra != -1 {
						y2 = 0
						for y := downgraded.Bounds().Min.Y; y < downgraded.Bounds().Max.Y; y++ {
							x2 := 0
							for x := downgraded.Bounds().Max.X - nbPixels; x < downgraded.Bounds().Max.X; x++ {
								im.Set(x2, y2, downgraded.At(x, y))
								x2++
							}
							y2++
						}
					}
					newFilename := strconv.Itoa(i) + strings.TrimSuffix(filename, path.Ext(filename)) + ".png"
					fmt.Fprintf(os.Stdout, "Saving downgraded image iteration (%d) into (%s)\n", i, newFilename)
					gfx.Png(*output+string(filepath.Separator)+newFilename, im)
					fmt.Fprintf(os.Stdout, "Tranform image in sprite iteration (%d)\n", i)
					gfx.SpriteTransform(im, newPalette, size, screenMode, newFilename, *output, *noAmsdosHeader, *plusMode)
				}
			}
		}
		if *keephigh != -1 || *losthigh != -1 {
			fmt.Fprintf(os.Stdout, "keephigh/losthigh: Iterations (%d)\n", *iterations)
			for i := 0; i < *iterations; i++ {
				nbPixels := 0
				if *keephigh != -1 {
					nbPixels = (*keephigh * (1 + i))
				} else {
					if *losthigh != -1 {
						nbPixels = (*losthigh * (1 + i))
					}
				}
				im := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{downgraded.Bounds().Max.X, downgraded.Bounds().Max.Y}})
				y2 := 0
				for y := downgraded.Bounds().Min.Y + nbPixels; y < downgraded.Bounds().Max.Y; y++ {
					x2 := 0
					for x := downgraded.Bounds().Min.X; x < downgraded.Bounds().Max.X; x++ {
						im.Set(x2, y2, downgraded.At(x, y))
						x2++
					}
					y2++
				}
				if *keephigh != -1 {
					for y := downgraded.Bounds().Min.Y; y < nbPixels; y++ {
						x2 := 0
						for x := downgraded.Bounds().Min.X; x < downgraded.Bounds().Max.X; x++ {
							im.Set(x2, y2, downgraded.At(x, y))
							x2++
						}
						y2++
					}
				}
				newFilename := strconv.Itoa(i) + strings.TrimSuffix(filename, path.Ext(filename)) + ".png"
				fmt.Fprintf(os.Stdout, "Saving downgraded image iteration (%d) into (%s)\n", i, newFilename)
				gfx.Png(*output+string(filepath.Separator)+newFilename, im)
				fmt.Fprintf(os.Stdout, "Tranform image in sprite iteration (%d)\n", i)
				gfx.SpriteTransform(im, newPalette, size, screenMode, newFilename, *output, *noAmsdosHeader, *plusMode)
			}
		} else {
			if *keeplow != -1 || *lostlow != -1 {
				fmt.Fprintf(os.Stdout, "keeplow/lostlow: Iterations (%d)\n", *iterations)
				for i := 0; i < *iterations; i++ {
					nbPixels := 0
					if *keeplow != -1 {
						nbPixels = (*keeplow * (1 + i))
					} else {
						if *lostlow != -1 {
							nbPixels = (*lostlow * (1 + i))
						}
					}
					im := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{downgraded.Bounds().Max.X, downgraded.Bounds().Max.Y}})
					y2 := downgraded.Bounds().Max.Y - 1
					for y := downgraded.Bounds().Max.Y - nbPixels; y >= downgraded.Bounds().Min.Y; y-- {
						x2 := 0
						for x := downgraded.Bounds().Min.X; x < downgraded.Bounds().Max.X; x++ {
							im.Set(x2, y2, downgraded.At(x, y))
							x2++
						}
						y2--
					}
					if *keeplow != -1 {
						for y := downgraded.Bounds().Max.Y - 1; y >= downgraded.Bounds().Max.Y-nbPixels; y-- {
							x2 := 0
							for x := downgraded.Bounds().Min.X; x < downgraded.Bounds().Max.X; x++ {
								im.Set(x2, y2, downgraded.At(x, y))
								x2++
							}
							y2--
						}
					}
					newFilename := strconv.Itoa(i) + strings.TrimSuffix(filename, path.Ext(filename)) + ".png"
					fmt.Fprintf(os.Stdout, "Saving downgraded image iteration (%d) into (%s)\n", i, newFilename)
					gfx.Png(*output+string(filepath.Separator)+newFilename, im)
					fmt.Fprintf(os.Stdout, "Tranform image in sprite iteration (%d)\n", i)
					gfx.SpriteTransform(im, newPalette, size, screenMode, newFilename, *output, *noAmsdosHeader, *plusMode)
				}
			}
		}
	} else {
		if !customDimension {
			gfx.Transform(downgraded, newPalette, size, *picturePath, *output, *noAmsdosHeader, *plusMode)
		} else {
			fmt.Fprintf(os.Stdout, "Transform image in sprite.\n")
			gfx.SpriteTransform(downgraded, newPalette, size, screenMode, filename, *output, *noAmsdosHeader, *plusMode)
		}
	}
}

package resizer

import "gopkg.in/gographics/imagick.v2/imagick"

type Resizer struct {
}

func New() Resizer {
	return Resizer{}
}

func (r *Resizer) Resize(width, height uint, inputFilename, outputFilename string) error {

	imagick.Initialize()
	// Schedule cleanup
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()

	err := mw.ReadImage(inputFilename)
	if err != nil {
		panic(err)
	}

	// Resize the image using the Lanczos filter
	// The blur factor is a float, where > 1 is blurry, < 1 is sharp
	err = mw.ResizeImage(width, height, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		panic(err)
	}

	// Set the compression quality to 95 (high quality = low compression)
	err = mw.SetImageCompressionQuality(95)
	if err != nil {
		panic(err)
	}

	err = mw.WriteImage(outputFilename)
	if err != nil {
		panic(err)
	}

	return nil
}

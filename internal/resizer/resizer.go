package resizer

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/gographics/imagick.v2/imagick"
)

type Resizer struct{}

// New is a resizer constructor.
func New() *Resizer {
	return &Resizer{}
}

var (
	ErrFileRead       = errors.New("unable to read a file")
	ErrImageResize    = errors.New("unable to resize an image")
	ErrQualitySetting = errors.New("unable to set a compression quality")
)

// Resize modifies file sizes by given slice of bytes.
// Note that Resize upscales file if source file is smaller!
func (r *Resizer) Resize(width, height uint, image []byte) ([]byte, error) {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	err := mw.ReadImageBlob(image)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileRead, err)
	}

	err = mw.ResizeImage(width, height, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrImageResize, err)
	}

	err = mw.SetImageCompressionQuality(95)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrQualitySetting, err)
	}

	return mw.GetImageBlob(), nil
}

package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v2/config"
	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/imagetype"
	"github.com/imgproxy/imgproxy/v2/imath"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/vips"
)

func canScaleOnLoad(imgtype imagetype.Type, scale float64) bool {
	if imgtype == imagetype.SVG {
		return true
	}

	if config.DisableShrinkOnLoad || scale >= 1 {
		return false
	}

	return imgtype == imagetype.JPEG || imgtype == imagetype.WEBP
}

func calcJpegShink(scale float64, imgtype imagetype.Type) int {
	shrink := int(1.0 / scale)

	switch {
	case shrink >= 8:
		return 8
	case shrink >= 4:
		return 4
	case shrink >= 2:
		return 2
	}

	return 1
}

func scaleOnLoad(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	prescale := math.Max(pctx.wscale, pctx.hscale)

	if pctx.trimmed || prescale == 1 || imgdata == nil || !canScaleOnLoad(pctx.imgtype, prescale) {
		return nil
	}

	jpegShrink := calcJpegShink(prescale, pctx.imgtype)

	if pctx.imgtype == imagetype.JPEG && jpegShrink == 1 {
		return nil
	}

	if err := img.Load(imgdata, jpegShrink, prescale, 1); err != nil {
		return err
	}

	// Update scales after scale-on-load
	newWidth, newHeight, _, _ := extractMeta(img, po.Rotate, po.AutoRotate)

	pctx.wscale = float64(pctx.srcWidth) * pctx.wscale / float64(newWidth)
	if pctx.srcWidth == imath.Scale(pctx.srcWidth, pctx.wscale) {
		pctx.wscale = 1.0
	}

	pctx.hscale = float64(pctx.srcHeight) * pctx.hscale / float64(newHeight)
	if pctx.srcHeight == imath.Scale(pctx.srcHeight, pctx.hscale) {
		pctx.hscale = 1.0
	}

	return nil
}
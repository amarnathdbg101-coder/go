package filekit

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
)

func init() {
	// Register standard image formats for image.Decode
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "GIF8?a", gif.Decode, gif.DecodeConfig)
}

// ImageInfo holds metadata of a processed image.
type ImageInfo struct {
	Width  int
	Height int
	Format string
}

// DecodeImage decodes an image stream into standard image.Image interface.
func DecodeImage(r io.Reader) (image.Image, string, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	return img, format, nil
}

// DecodeImageConfig decodes only image header metadata without decompressing pixel data.
// Fast pre-flight check to detect format and prevent pixel-flood / decompression bomb attacks.
func DecodeImageConfig(r io.Reader) (image.Config, string, error) {
	cfg, format, err := image.DecodeConfig(r)
	if err != nil {
		return image.Config{}, "", fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	return cfg, format, nil
}

// CalculateAspectFit calculates proportional target dimensions maintaining aspect ratio.
// If original image is already smaller than bounds, original dimensions are preserved (no upscaling).
func CalculateAspectFit(origW, origH, maxW, maxH int) (int, int) {
	if origW <= 0 || origH <= 0 {
		return 0, 0
	}
	if maxW <= 0 || maxH <= 0 {
		return origW, origH
	}

	if origW <= maxW && origH <= maxH {
		return origW, origH
	}

	scaleW := float64(maxW) / float64(origW)
	scaleH := float64(maxH) / float64(origH)
	scale := math.Min(scaleW, scaleH)

	targetW := int(math.Round(float64(origW) * scale))
	targetH := int(math.Round(float64(origH) * scale))

	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}

	return targetW, targetH
}

// ResizeImage resizes an image to target dimensions using pure Go Bilinear interpolation.
// No external Cgo/libvips dependencies required — builds and runs seamlessly everywhere.
func ResizeImage(src image.Image, targetW, targetH int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == targetW && srcH == targetH {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))

	xRatio := float64(srcW-1) / float64(targetW)
	yRatio := float64(srcH-1) / float64(targetH)

	for y := 0; y < targetH; y++ {
		srcY := float64(y) * yRatio
		yLow := int(srcY)
		yHigh := yLow + 1
		if yHigh >= srcH {
			yHigh = yLow
		}
		yWeight := srcY - float64(yLow)

		for x := 0; x < targetW; x++ {
			srcX := float64(x) * xRatio
			xLow := int(srcX)
			xHigh := xLow + 1
			if xHigh >= srcW {
				xHigh = xLow
			}
			xWeight := srcX - float64(xLow)

			c00 := src.At(bounds.Min.X+xLow, bounds.Min.Y+yLow)
			c10 := src.At(bounds.Min.X+xHigh, bounds.Min.Y+yLow)
			c01 := src.At(bounds.Min.X+xLow, bounds.Min.Y+yHigh)
			c11 := src.At(bounds.Min.X+xHigh, bounds.Min.Y+yHigh)

			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Bilinear interpolation formula
			topR := float64(r00)*(1-xWeight) + float64(r10)*xWeight
			topG := float64(g00)*(1-xWeight) + float64(g10)*xWeight
			topB := float64(b00)*(1-xWeight) + float64(b10)*xWeight
			topA := float64(a00)*(1-xWeight) + float64(a10)*xWeight

			botR := float64(r01)*(1-xWeight) + float64(r11)*xWeight
			botG := float64(g01)*(1-xWeight) + float64(g11)*xWeight
			botB := float64(b01)*(1-xWeight) + float64(b11)*xWeight
			botA := float64(a01)*(1-xWeight) + float64(a11)*xWeight

			finalR := uint8((topR*(1-yWeight) + botR*yWeight) / 257)
			finalG := uint8((topG*(1-yWeight) + botG*yWeight) / 257)
			finalB := uint8((topB*(1-yWeight) + botB*yWeight) / 257)
			finalA := uint8((topA*(1-yWeight) + botA*yWeight) / 257)

			dst.Set(x, y, color.RGBA{R: finalR, G: finalG, B: finalB, A: finalA})
		}
	}

	return dst
}

// CompressAndEncode encodes the image to specified format.
// For JPEG, compresses with quality (e.g. 80-85) to significantly reduce storage footprint.
func CompressAndEncode(img image.Image, format string, jpegQuality int) ([]byte, error) {
	if jpegQuality <= 0 || jpegQuality > 100 {
		jpegQuality = 85
	}

	var buf bytes.Buffer

	switch format {
	case "jpeg", "jpg":
		opts := &jpeg.Options{Quality: jpegQuality}
		if err := jpeg.Encode(&buf, img, opts); err != nil {
			return nil, err
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	case "gif":
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, err
		}
	default:
		// Fallback to high quality JPEG
		opts := &jpeg.Options{Quality: jpegQuality}
		if err := jpeg.Encode(&buf, img, opts); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

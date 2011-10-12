// Package pnm implements a decoder for PBM, PGM and PPM files.
//
// Specifications can be found at http://netpbm.sourceforge.net/doc/#formats
// PAM files are currently not supported.
package pnm

import (
	"os"
	"io"
	"fmt"
	"bufio"
	"unicode"
	"image"
	"image/color"
)

type PNMConfig struct {
	Width  int
	Height int
	Maxval int
	magic  string
}

func decodePlainBW(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewGray(image.Rect(0, 0, c.Width, c.Height))
	pixelCount := len(m.Pix)

	for i := 0; i < pixelCount; i++ {
		if _, err := fmt.Fscan(r, &m.Pix[i]); err != nil {
			return nil, err
		}
		if m.Pix[i] == 0 {
			m.Pix[i] = 255
		} else {
			m.Pix[i] = 0
		}
	}

	return m, nil
}

func decodePlainGray(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewGray(image.Rect(0, 0, c.Width, c.Height))
	pixelCount := len(m.Pix)

	for i := 0; i < pixelCount; i++ {
		if _, err := fmt.Fscan(r, &m.Pix[i]); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func decodePlainGray16(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewGray16(image.Rect(0, 0, c.Width, c.Height))
	var col uint16

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if _, err := fmt.Fscan(r, &col); err != nil {
				return nil, err
			}
			m.Set(x, y, color.Gray16{col})
		}
	}

	return m, nil
}

func decodePlainRGB(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	count := len(m.Pix)

	for i := 0; i < count; i += 4 {
		if _, err := fmt.Fscan(r, &m.Pix[i]); err != nil {
			return nil, err
		}
		if _, err := fmt.Fscan(r, &m.Pix[i+1]); err != nil {
			return nil, err
		}
		if _, err := fmt.Fscan(r, &m.Pix[i+2]); err != nil {
			return nil, err
		}
		m.Pix[i+3] = 0xff
	}

	return m, nil
}

func decodePlainRGB64(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewRGBA64(image.Rect(0, 0, c.Width, c.Height))
	var cr, cg, cb uint16

	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			if _, err := fmt.Fscan(r, &cr); err != nil {
				return nil, err
			}
			if _, err := fmt.Fscan(r, &cg); err != nil {
				return nil, err
			}
			if _, err := fmt.Fscan(r, &cb); err != nil {
				return nil, err
			}
			m.Set(x, y, color.RGBA64{cr, cg, cb, 0xffff})
		}
	}

	return m, nil
}

// unpackByte unpacks 8 one bit pixels from byte b into slice bit.
//
// The bits are unpacked such that the most significant bit becomes the
// first value in the slice. If there are less than 8 values in bit,the
// remaining bits are ignored. If there are more than 8 values in bit,
// these remain unchanged.
func unpackByte(bit []uint8, b byte) {
	n := len(bit)
	if n > 8 {
		n = 8
	}
	for i := 0; i < n; i++ {
		if b&128 == 0 {
			bit[i] = 255
		}
		b = b << 1
	}
}

func decodeRawBW(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewGray(image.Rect(0, 0, c.Width, c.Height))

	byteCount := c.Width / 8
	if c.Width%8 != 0 {
		byteCount += 1
	}
	row := make([]byte, byteCount)
	pos := 0

	for y := 0; y < c.Height; y++ {
		if _, err := io.ReadFull(r, row); err != nil {
			return nil, err
		}
		bitsLeft := c.Width
		for _, b := range row {
			n := bitsLeft
			if n > 8 {
				n = 8
			}
			unpackByte(m.Pix[pos:pos+n], b)
			bitsLeft -= n
			pos += n
		}
	}

	return m, nil
}

func decodeRawGray(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewGray(image.Rect(0, 0, c.Width, c.Height))
	_, err := io.ReadFull(r, m.Pix)
	return m, err
}

func decodeRawGray16(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewGray16(image.Rect(0, 0, c.Width, c.Height))
	_, err := io.ReadFull(r, m.Pix)
	return m, err
}

func decodeRawRGB(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	count := len(m.Pix)
	for i := 0; i < count; i += 4 {
		pixel := m.Pix[i : i+3]
		m.Pix[i+3] = 0xff

		if _, err := io.ReadFull(r, pixel); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func decodeRawRGB64(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	m := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	count := len(m.Pix)

	for i := 0; i < count; i += 8 {
		pixel := m.Pix[i : i+6]
		m.Pix[i+6] = 0xff
		m.Pix[i+7] = 0xff

		if _, err := io.ReadFull(r, pixel); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func decodePAM(r io.Reader, c PNMConfig) (image.Image, os.Error) {
	return nil, os.NewError("pnm: reading PAM images is not supported yet.")
}

// Decode reads a PNM image from r and returns it as an image.Image.
//
// The type of Image returned depends on the PNM contents:
//  - PBM: image.Gray with black = 0 and white = 255
//  - PGM: image.Gray or image.Gray16, values as in the file
//  - PPM: image.RGBA or image.RGBA64, values as in the file
//  - PAM: not supported (yet)
func Decode(r io.Reader) (image.Image, os.Error) {
	br := bufio.NewReader(r)
	c, err := DecodeConfigPNM(br)

	if err != nil {
		err = fmt.Errorf("pnm: parsing header failed: %v", err)
		return nil, err
	}

	switch c.magic {
	case "P1":
		return decodePlainBW(br, c)
	case "P2":
		if c.Maxval < 256 {
			return decodePlainGray(br, c)
		} else {
			return decodePlainGray16(br, c)
		}
	case "P3":
		if c.Maxval < 256 {
			return decodePlainRGB(br, c)
		} else {
			return decodePlainRGB64(br, c)
		}
	case "P4":
		return decodeRawBW(br, c)
	case "P5":
		if c.Maxval < 256 {
			return decodeRawGray(br, c)
		} else {
			return decodeRawGray16(br, c)
		}
	case "P6":
		if c.Maxval < 256 {
			return decodeRawRGB(br, c)
		} else {
			return decodeRawRGB64(br, c)
		}
	case "P7":
		return decodePAM(br, c)
	}

	return nil, fmt.Errorf("pnm: could not decode, invalid magic value %s", c.magic[0:2])
}

// skipComments skips all comments (and whitespace) that may occur between PNM
// header tokens.
//
// The singleSpace argument is used to scan comments between the header and the
// raster data where only a single whitespace delimiter is allowed. This
// prevents scanning the image data.
func skipComments(r *bufio.Reader, singleSpace bool) (err os.Error) {
	for {
		// Skip whitespace
		c, err := r.ReadByte()
		for unicode.IsSpace(int(c)) {
			if c, err = r.ReadByte(); err != nil {
				return err
			}
			if singleSpace {
				break
			}
		}
		// If there are no more comments, unread the last byte and return.
		if c != '#' {
			r.UnreadByte()
			return nil
		}
		// A comment ends with a newline or carriage return.
		for c != '\n' && c != '\r' {
			if c, err = r.ReadByte(); err != nil {
				return
			}
		}
	}
	return
}

// DecodeConfigPNM reads the header data of PNM files.
//
// In contrast to DecodeConfig it returns a PNMConfig struct that contains
// some PNM specific information that may be needed for reading the image
// or when applying (gamma) color corrections (which is not implemented).
func DecodeConfigPNM(r *bufio.Reader) (c PNMConfig, err os.Error) {
	// PNM magic number
	if _, err = fmt.Fscan(r, &c.magic); err != nil {
		return
	}
	switch c.magic {
	case "P1", "P2", "P3", "P4", "P5", "P6":
	case "P7":
		return c, os.NewError("pnm: reading PAM images is not supported (yet).")
	default:
		return c, os.NewError("pnm: invalid format " + c.magic[0:2])
	}

	// Image width
	if err = skipComments(r, false); err != nil {
		return
	}
	if _, err = fmt.Fscan(r, &c.Width); err != nil {
		return c, os.NewError("pnm: could not read image width, " + err.String())
	}
	// Image height
	if err = skipComments(r, false); err != nil {
		return
	}
	if _, err = fmt.Fscan(r, &c.Height); err != nil {
		return c, os.NewError("pnm: could not read image height, " + err.String())
	}
	// Number of colors, only for gray and color images.
	// For black and white images this is 2, obviously.
	if c.magic == "P1" || c.magic == "P4" {
		c.Maxval = 2
	} else {
		if err = skipComments(r, false); err != nil {
			return
		}
		if _, err = fmt.Fscan(r, &c.Maxval); err != nil {
			return c, os.NewError("pnm: could not read number of colors, " + err.String())
		}
	}

	if c.Maxval > 65535 || c.Maxval <= 0 {
		err = fmt.Errorf("pnm: maximum depth is 16 bit (65,535) colors but %d colors found", c.Maxval)
		return
	}

	// Skip comments after header.
	if err = skipComments(r, true); err != nil {
		return
	}

	return c, nil
}

// DecodeConfig returns the color model and dimensions of a PNM image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, os.Error) {
	br := bufio.NewReader(r)
	c, err := DecodeConfigPNM(br)
	if err != nil {
		return image.Config{}, err
	}

	var cm color.Model
	switch c.magic {
	case "P1", "P4":
		cm = color.GrayModel
	case "P2", "P5":
		if c.Maxval < 256 {
			cm = color.GrayModel
		} else {
			cm = color.Gray16Model
		}
	case "P3", "P6":
		if c.Maxval < 256 {
			cm = color.RGBAModel
		} else {
			cm = color.RGBA64Model
		}
	}

	return image.Config{cm, c.Width, c.Height}, nil
}

func init() {
	image.RegisterFormat("pbm plain", "P1", Decode, DecodeConfig)
	image.RegisterFormat("pgm plain", "P2", Decode, DecodeConfig)
	image.RegisterFormat("ppm plain", "P3", Decode, DecodeConfig)
	image.RegisterFormat("pbm raw", "P4", Decode, DecodeConfig)
	image.RegisterFormat("pgm raw", "P5", Decode, DecodeConfig)
	image.RegisterFormat("ppm raw", "P6", Decode, DecodeConfig)
	//image.RegisterFormat("pam", "P7", Decode, DecodeConfig)
}

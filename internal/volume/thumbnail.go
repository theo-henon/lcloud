package volume

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const thumbnailMaxEdge = 256

type ThumbnailGenerator struct{}

func NewThumbnailGenerator() *ThumbnailGenerator {
	return &ThumbnailGenerator{}
}

func (g *ThumbnailGenerator) thumbnailsDir(rootPath string) string {
	return filepath.Join(rootPath, CacheDir, ThumbnailsDir)
}

func (g *ThumbnailGenerator) ThumbnailPath(rootPath, fileID string) string {
	return filepath.Join(g.thumbnailsDir(rootPath), fileID+".jpg")
}

func (g *ThumbnailGenerator) Generate(rootPath, fileID string, reader io.Reader) (string, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return "", err
	}

	thumb := resizeThumbnail(img)
	if err := os.MkdirAll(g.thumbnailsDir(rootPath), 0o755); err != nil {
		return "", err
	}

	outPath := g.ThumbnailPath(rootPath, fileID)
	file, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := jpeg.Encode(file, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return "", err
	}
	return outPath, nil
}

func (g *ThumbnailGenerator) Open(rootPath, fileID string) (*os.File, error) {
	return os.Open(g.ThumbnailPath(rootPath, fileID))
}

func (g *ThumbnailGenerator) Delete(rootPath, fileID string) error {
	err := os.Remove(g.ThumbnailPath(rootPath, fileID))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func resizeThumbnail(src image.Image) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	scale := float64(thumbnailMaxEdge) / float64(max(width, height))
	if scale >= 1 {
		return src
	}

	newW := max(1, int(float64(width)*scale))
	newH := max(1, int(float64(height)*scale))
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func decodeImageFormat(data []byte) (string, bool) {
	if len(data) < 12 {
		return "", false
	}
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", true
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		return "image/png", true
	case bytes.HasPrefix(data, []byte("GIF87a")), bytes.HasPrefix(data, []byte("GIF89a")):
		return "image/gif", true
	case bytes.HasPrefix(data, []byte("RIFF")) && bytes.Contains(data[:12], []byte("WEBP")):
		return "image/webp", true
	default:
		return "", false
	}
}

func isImageMime(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func registerImageFormats() {
	image.RegisterFormat("jpeg", "\xff\xd8\xff", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)
}

func init() {
	registerImageFormats()
}

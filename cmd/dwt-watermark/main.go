package main

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/yyyoichi/watermark_zero"
	"github.com/yyyoichi/watermark_zero/mark"
)

var CLI struct {
	Embed   EmbedCmd   `cmd:"" help:"Embed watermark into image"`
	Extract ExtractCmd `cmd:"" help:"Extract watermark from image"`
	Detect  DetectCmd  `cmd:"" help:"Detect if image has watermark"`
}

type EmbedCmd struct {
	Input       string `arg:"" required help:"Input image file (PNG/JPEG)"`
	Output      string `arg:"" required help:"Output image file"`
	Watermark   string `short:"w" required help:"Watermark text to embed"`
	BlockWidth  int    `short:"b" default:"8" help:"Block width"`
	BlockHeight int    `short:"e" default:"6" help:"Block height"`
	D1          int    `short:"1" default:"36" help:"D1 parameter (embedding strength)"`
	D2          int    `short:"2" default:"20" help:"D2 parameter (secondary strength)"`
	Quality     int    `short:"q" default:"95" help:"JPEG quality if output is JPEG"`
}

func (e *EmbedCmd) Run() error {
	img, err := loadImage(e.Input)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	w, err := watermark.New(
		watermark.WithBlockShape(e.BlockWidth, e.BlockHeight),
		watermark.WithD1D2(e.D1, e.D2),
	)
	if err != nil {
		return fmt.Errorf("failed to create watermark instance: %w", err)
	}

	m := mark.NewString(e.Watermark)
	fmt.Printf("Embedding watermark: %s\n", e.Watermark)
	fmt.Printf("Extract size: %d bits\n", m.ExtractSize())

	markedImg, err := w.Embed(context.Background(), img, m)
	if err != nil {
		return fmt.Errorf("failed to embed watermark: %w", err)
	}

	if err := saveImage(e.Output, markedImg, e.Quality); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Printf("Watermark embedded successfully: %s\n", e.Output)
	fmt.Printf("Note: Use --size %d to extract this watermark\n", m.ExtractSize())
	return nil
}

type ExtractCmd struct {
	Input       string `arg:"" required help:"Image file to extract watermark from"`
	Output      string `short:"o" help:"Output file for extracted watermark text"`
	Size        int    `short:"s" default:"256" help:"Expected watermark size in bits (use the value from embed output)"`
	BlockWidth  int    `short:"b" default:"8" help:"Block width (must match embed settings)"`
	BlockHeight int    `short:"e" default:"6" help:"Block height (must match embed settings)"`
	D1          int    `short:"1" default:"36" help:"D1 parameter (must match embed settings)"`
	D2          int    `short:"2" default:"20" help:"D2 parameter (must match embed settings)"`
}

func (e *ExtractCmd) Run() error {
	img, err := loadImage(e.Input)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	w, err := watermark.New(
		watermark.WithBlockShape(e.BlockWidth, e.BlockHeight),
		watermark.WithD1D2(e.D1, e.D2),
	)
	if err != nil {
		return fmt.Errorf("failed to create watermark instance: %w", err)
	}

	exM := mark.NewExtract(e.Size)
	decoded, err := w.Extract(context.Background(), img, exM)
	if err != nil {
		return fmt.Errorf("failed to extract watermark: %w", err)
	}

	text := decoded.DecodeToString()
	text = strings.TrimRight(text, "\x00")
	if text == "" {
		fmt.Println("No watermark detected")
		return nil
	}

	if e.Output != "" {
		if err := os.WriteFile(e.Output, []byte(text), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Watermark: %s\n", text)
		fmt.Printf("Saved to: %s\n", e.Output)
	} else {
		fmt.Printf("Watermark: %s\n", text)
	}

	return nil
}

type DetectCmd struct {
	Input       string `arg:"" required help:"Image file to detect watermark"`
	Size        int    `short:"s" default:"256" help:"Expected watermark size in bits"`
	BlockWidth  int    `short:"b" default:"8" help:"Block width"`
	BlockHeight int    `short:"e" default:"6" help:"Block height"`
	D1          int    `short:"1" default:"36" help:"D1 parameter"`
	D2          int    `short:"2" default:"20" help:"D2 parameter"`
}

func (d *DetectCmd) Run() error {
	img, err := loadImage(d.Input)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	w, err := watermark.New(
		watermark.WithBlockShape(d.BlockWidth, d.BlockHeight),
		watermark.WithD1D2(d.D1, d.D2),
	)
	if err != nil {
		return fmt.Errorf("failed to create watermark instance: %w", err)
	}

	exM := mark.NewExtract(d.Size)
	decoded, err := w.Extract(context.Background(), img, exM)
	if err != nil {
		fmt.Println("No watermark detected")
		return nil
	}

	text := decoded.DecodeToString()
	text = strings.TrimRight(text, "\x00")
	text = strings.TrimSpace(text)

	if text == "" || !isLikelyWatermark(text) {
		fmt.Println("No watermark detected")
		return nil
	}

	fmt.Printf("Watermark detected: %s\n", text)
	return nil
}

func isLikelyWatermark(s string) bool {
	if len(s) == 0 {
		return false
	}

	printable := 0
	for _, r := range s {
		if r >= 32 && r < 127 {
			printable++
		}
	}

	return float64(printable)/float64(len(s)) > 0.7
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func saveImage(path string, img image.Image, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := strings.ToLower(path[strings.LastIndex(path, ".")+1:])
	switch ext {
	case "jpg", "jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
	default:
		return png.Encode(f, img)
	}
}

func main() {
	ctx := kong.Parse(&CLI)
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

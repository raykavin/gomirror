package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type ScreenSize struct {
	Width  uint16
	Height uint16
}

type ScreenPoint struct {
	X int32
	Y int32
}

func main() {
	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Vídeo Capturado",
		Bounds: pixel.R(0, 0, 400, 855),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var (
		sprite         *pixel.Sprite
		prevScreen     image.Image
		screenWidth    = float64(win.Bounds().W())
		screenHeight   = float64(win.Bounds().H())
		screenOffsetX  = win.Bounds().Min.X
		screenOffsetY  = win.Bounds().Min.Y
		screenToWindow = pixel.IM.Moved(win.Bounds().Center())
		screenSize     ScreenSize
	)

	for !win.Closed() {
		img, err := captureScreen()
		if err != nil {
			log.Fatal(err)
		}

		img = imaging.Resize(img, int(screenWidth), int(screenHeight), imaging.Lanczos)

		if prevScreen == nil || !imagesEqual(prevScreen, img) {
			prevScreen = img

			pic := pixel.PictureDataFromImage(img)
			if sprite == nil {
				sprite = pixel.NewSprite(pic, pic.Bounds())
			} else {
				sprite.Set(pic, pic.Bounds())
			}

			win.Clear(colornames.Black)
			sprite.Draw(win, screenToWindow)
			win.Update()

			if win.JustPressed(pixelgl.MouseButtonLeft) {
				mousePos := win.MousePosition()
				clickX := int(mousePos.X - screenOffsetX)
				clickY := int(mousePos.Y - screenOffsetY)

				screenX, screenY := convertToScreenCoordinates(clickX, clickY, screenSize.Width, screenSize.Height)

				err := tapScreen(screenX, screenY)
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func captureScreen() (image.Image, error) {
	cmd := exec.Command("adb", "exec-out", "screencap", "-p")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(string(output), "\x89PNG") {
		return nil, fmt.Errorf("erro ao receber a captura de tela")
	}

	img, err := png.Decode(bytes.NewReader(output))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// func getWindowSize() (ScreenSize, error) {
// 	cmd := exec.Command("adb", "shell", "wm", "size")
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return ScreenSize{}, err
// 	}

// 	strOutput := strings.TrimPrefix(string(output), "Physical size: ")
// 	strOutput = strings.TrimSuffix(strOutput, "\n")

// 	windowSize := strings.Split(strOutput, "x")

// 	width, err := strconv.ParseUint(windowSize[0], 10, 16)
// 	if err != nil {
// 		return ScreenSize{}, err
// 	}
// 	height, err := strconv.ParseUint(windowSize[1], 10, 16)
// 	if err != nil {
// 		return ScreenSize{}, err
// 	}

// 	return ScreenSize{
// 		Width:  uint16(width),
// 		Height: uint16(height),
// 	}, nil

// }

func convertToScreenCoordinates(clickX, clickY int, screenWidth, screenHeight uint16) (int, int) {
	screenWidthFloat := float64(screenWidth)
	screenHeightFloat := float64(screenHeight)

	screenX := int((float64(clickX) / screenWidthFloat) * screenWidthFloat)
	screenY := int((float64(clickY) / screenHeightFloat) * screenHeightFloat)

	return screenX, screenY
}

func tapScreen(x, y int) error {
	cmd := exec.Command("adb", "shell", "input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func imagesEqual(img1, img2 image.Image) bool {
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()
	if bounds1 != bounds2 {
		return false
	}
	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			if img1.At(x, y) != img2.At(x, y) {
				return false
			}
		}
	}

	return true
}

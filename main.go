package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
)

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func loadSprite(path string) (pixel.Sprite, error) {
	pic, err := loadPicture(path)
	if err != nil {
		return *pixel.NewSprite(pic, pic.Bounds()), err
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())
	return *sprite, nil
}

func loadTTF(path string, size float64, origin pixel.Vec) *text.Text {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	face := truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	})

	atlas := text.NewAtlas(face, text.ASCII)

	txt := text.New(origin, atlas)

	return txt

}

type mode int

const (
	menu   mode = 0
	flying mode = 1
	mapper mode = 2
)

func run() {
	// Set up window configs
	cfg := pixelgl.WindowConfig{ // Default: 1024 x 768
		Title:  "woosh!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
		//Monitor: pixelgl.PrimaryMonitor(),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Importantr variables
	var shipX, shipY, velX, velY, radians float64
	engineOn := false
	gravity := 0.02 // Default: 0.004
	shipAcc := 0.08 // Default: 0.008
	tilt := 0.025   // Default: 0.001
	whichOn := false
	onNumber := 0
	shipOffName := "ship.png"
	shipOn1Name := "ship_on.png"
	shipOn2Name := "ship_on2.png"
	camVector := win.Bounds().Center()

	bg, _ := loadSprite("sky.png")

	// Jetpack - Rendering
	engineOff, err := loadSprite(shipOffName)
	if err != nil {
		panic(err)
	}
	engineOn1, err := loadSprite(shipOn1Name)
	if err != nil {
		panic(err)
	}
	engineOn2, err := loadSprite(shipOn2Name)
	if err != nil {
		panic(err)
	}

	txt := loadTTF("oldgamefatty.ttf", 50, pixel.V(win.Bounds().Center().X-450, win.Bounds().Center().Y-200))

	currentSprite := engineOff

	frameCounter := 0
	areadout := int(shipY)
	vreadout := int(velY)

	currentmode := flying

	// Game Loop
	for !win.Closed() {

		if currentmode == flying {

			win.Update()
			win.Clear(colornames.Green)
			// Jetpack - Controls
			engineOn = win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW)

			if win.Pressed(pixelgl.KeyRight) || win.Pressed(pixelgl.KeyD) {
				radians -= tilt
			} else if win.Pressed(pixelgl.KeyLeft) || win.Pressed(pixelgl.KeyA) {
				radians += tilt
			} else if win.JustPressed(pixelgl.KeyW) {
				currentmode = menu
			} else if win.JustPressed(pixelgl.KeyM) {
				currentmode = mapper
			}

			if shipY < 0 {
				shipY = 0
				velY = -0.3 * velY
			}

			if engineOn {

				heading := pixel.Unit(radians)

				acc := heading.Scaled(shipAcc)

				velY += acc.X
				velX -= acc.Y

				whichOn = !whichOn
				onNumber++
				if onNumber == 5 { // every 5 frames, toggle anishipMation
					onNumber = 0
					if whichOn {
						currentSprite = engineOn1
					} else {
						currentSprite = engineOn2
					}
				}
			} else {
				currentSprite = engineOff
			}

			velY -= gravity

			positionVector := pixel.V(win.Bounds().Center().X+shipX, win.Bounds().Center().Y+shipY-372)
			shipMat := pixel.IM
			shipMat = shipMat.Scaled(pixel.ZV, 4)
			shipMat = shipMat.Moved(positionVector)
			shipMat = shipMat.Rotated(positionVector, radians)

			shipX += velX
			shipY += velY

			// Camera
			camVector.X += (positionVector.X - camVector.X) * 0.2
			camVector.Y += (positionVector.Y - camVector.Y) * 0.2

			if camVector.X > 25085 {
				camVector.X = 25085
			} else if camVector.X < -14843 {
				camVector.X = -14843
			}

			if camVector.Y > 22500 {
				camVector.Y = 22500
			}

			cam := pixel.IM.Moved(win.Bounds().Center().Sub(camVector))

			win.SetMatrix(cam)

			// Drawing to the screen
			win.SetSmooth(true)
			bg.Draw(win, pixel.IM.Moved(pixel.V(win.Bounds().Center().X, win.Bounds().Center().Y+766)).Scaled(pixel.ZV, 10))

			//doesnt work
			//		txt.Draw(win, shipMat)
			txt.Clear()
			fmt.Fprintf(txt, "altitude: %d\n", areadout)
			fmt.Fprintf(txt, "velocity: %d", vreadout)
			txt.Draw(win, pixel.IM.Moved(positionVector))

			if frameCounter >= 10 {
				areadout = int(shipY)
				vreadout = int(velY)
				frameCounter = 0
			}

			win.SetSmooth(false)
			currentSprite.Draw(win, shipMat)

			frameCounter++
		}
		if currentmode == menu {
			currentitem := 0

			for currentmode == menu {
				win.Update()
				win.Clear(colornames.Blue)

				if win.JustPressed(pixelgl.KeyUp) {
					currentitem++
					if currentitem > 2 {
						currentitem = 0
					}
				}

				if win.JustPressed(pixelgl.KeyDown) {
					currentitem--
					if currentitem < 0 {
						currentitem = 2
					}
				}

				if win.JustPressed(pixelgl.KeyEnter) {

					switch currentitem {
					case 0:
						currentmode = flying
					case 1:
						win.Destroy()
					case 2:
						win.Destroy()
					}
				}

				win.SetSmooth(false)

				txt.Clear()
				fmt.Fprintf(txt, "%d\n", currentitem)
				txt.Draw(win, pixel.IM)
			}
		}

		if currentmode == mapper {
			win.Update()
			win.Clear(colornames.Pink)

			if win.JustPressed(pixelgl.KeyQ) || win.JustPressed(pixelgl.KeyM) {
				currentmode = flying
			}

			win.SetSmooth(false)
		}
	}
}

func main() {
	pixelgl.Run(run)
}

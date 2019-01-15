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
	Menu   mode = 0
	Flying mode = 1
	Map  mode = 2
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
	var jetX, jetY, velX, velY, radians float64
	jetpackOn := false
	gravity := 0.02 // Default: 0.004
	jetAcc := 0.08  // Default: 0.008
	tilt := 0.025   // Default: 0.001
	whichOn := false
	onNumber := 0
	jetpackOffName := "ship.png"
	jetpackOn1Name := "ship_on.png"
	jetpackOn2Name := "ship_on2.png"
	camVector := win.Bounds().Center()

	bg, _ := loadSprite("sky.png")

	// Jetpack - Rendering
	jetpackOff, err := loadSprite(jetpackOffName)
	if err != nil {
		panic(err)
	}
	jetpackOn1, err := loadSprite(jetpackOn1Name)
	if err != nil {
		panic(err)
	}
	jetpackOn2, err := loadSprite(jetpackOn2Name)
	if err != nil {
		panic(err)
	}

	txt := loadTTF("intuitive.ttf", 50, pixel.V(win.Bounds().Center().X-450, win.Bounds().Center().Y-200))

	currentSprite := jetpackOff

	frameCounter := 0
	areadout := int(jetY)
	vreadout := int(velY)

	currentmode := Flying

	// Game Loop
	for !win.Closed() {

		if currentmode == Flying {

		        win.Update()
			win.Clear(colornames.Green)
			// Jetpack - Controls
			jetpackOn = win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW)

			if win.Pressed(pixelgl.KeyRight) || win.Pressed(pixelgl.KeyD) {
				radians -= tilt
			} else if win.Pressed(pixelgl.KeyLeft) || win.Pressed(pixelgl.KeyA) {
				radians += tilt
			} else if win.JustPressed(pixelgl.KeyW) {
				currentmode = Menu
			} else if win.JustPressed(pixelgl.KeyM) {
				currentmode = Map
			}

			if jetY < 0 {
				jetY = 0
				velY = -0.3 * velY
			}

			if jetpackOn {

				heading := pixel.Unit(radians)

				acc := heading.Scaled(jetAcc)

				velY += acc.X
				velX -= acc.Y

				whichOn = !whichOn
				onNumber++
				if onNumber == 5 { // every 5 frames, toggle anijetMation
					onNumber = 0
					if whichOn {
						currentSprite = jetpackOn1
					} else {
						currentSprite = jetpackOn2
					}
				}
			} else {
				currentSprite = jetpackOff
			}

			velY -= gravity

			positionVector := pixel.V(win.Bounds().Center().X+jetX, win.Bounds().Center().Y+jetY-372)
			jetMat := pixel.IM
			jetMat = jetMat.Scaled(pixel.ZV, 4)
			jetMat = jetMat.Moved(positionVector)
			jetMat = jetMat.Rotated(positionVector, radians)

			jetX += velX
			jetY += velY

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
			//		txt.Draw(win, jetMat)
			txt.Clear()
			fmt.Fprintf(txt, "altitude: %d\n", areadout)
			fmt.Fprintf(txt, "velocity: %d", vreadout)
			txt.Draw(win, pixel.IM.Moved(positionVector))

			if frameCounter >= 10 {
				areadout = int(jetY)
				vreadout = int(velY)
				frameCounter = 0
			}

			win.SetSmooth(false)
			currentSprite.Draw(win, jetMat)

			frameCounter++
		}
		if currentmode == Menu {
		        win.Update()
			win.Clear(colornames.Blue)
			win.SetSmooth(false)

			if win.JustPressed(pixelgl.KeyQ) {
				currentmode = Flying
			}


		}


		if currentmode == Map {
		        win.Update()
			win.Clear(colornames.Pink)
			win.SetSmooth(false)

			if win.JustPressed(pixelgl.KeyQ) || win.JustPressed(pixelgl.KeyM) {
				currentmode = Flying
			}


		}



	}
}

func main() {
	pixelgl.Run(run)
}

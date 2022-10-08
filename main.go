package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var saxAudioController SaxAudioController

func main() {
	saxAudioController.notes = make(chan Note)
	saxAudioController.pause = make(chan int)
	saxAudioController.quit = make(chan int)
	StartSaxAudioPlayer(saxAudioController)
	defer func() { saxAudioController.quit <- 1 }()
	go func() {
		w := app.NewWindow()
		err := run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

var th = material.NewTheme(gofont.Collection())

var offColor = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
var onColor = color.NRGBA{R: 50, G: 50, B: 255, A: 255}

type Title struct {
	color color.NRGBA
}

func NewTitle() Title {
	var res Title
	res.color = offColor
	return res
}
func (self *Title) Layout(gtx layout.Context) layout.Dimensions {
	title := material.H1(th, "Hello, Gio")
	{ // handle input
		for _, ev := range gtx.Events(self) {
			log.Println("event: ", ev)
			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}
			switch e.Type {
			case pointer.Press:
				self.color = onColor
			case pointer.Release:
				self.color = offColor
			}
		}
	}
	title.Color = self.color
	title.Alignment = text.Middle
	dims := title.Layout(gtx)

	pointer.InputOp{Tag: self, Types: pointer.Press | pointer.Release}.Add(gtx.Ops)
	return dims
}

// ColorBox creates a widget with the specified dimensions and color.
func ColorBox(gtx layout.Context, size image.Point, color color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

var saxStateList = layout.List{Axis: layout.Vertical}

func SaxStateLayout(gtx layout.Context, state SaxState) layout.Dimensions {
	return saxStateList.Layout(gtx, len(saxButtonDrawingInstructions),
		func(gtx layout.Context, i int) layout.Dimensions {
			instr := saxButtonDrawingInstructions[i]
			return layout.UniformInset(unit.Dp(10)).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					var color color.NRGBA
					if *instr.sourceStateButton {
						color = onColor
					} else {
						color = offColor
					}
					return ColorBox(gtx,
						image.Point{int(instr.drawingSize * 60.0),
							int(instr.drawingSize * 60.0)}, color)
				})
		})
}

// return true if something changed
func updateSaxState(e key.Event) bool {
	touchedButton, ok := saxKeyMap[e.Name]
	if !ok {
		return false
	}
	changed := false
	if e.State == key.Press {
		if *touchedButton == false {
			*touchedButton = true
			changed = true
		}
	} else {
		if *touchedButton == true {
			*touchedButton = false
			changed = true
		}
	}
	return changed
}

func run(w *app.Window) error {
	var ops_ op.Ops
	ops := &ops_
	// title := NewTitle()
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(ops, e)
			// title.Layout(gtx)
			// handle keyboard input
			for _, ev := range gtx.Events(w) {
				e, ok := ev.(key.Event)
				if !ok {
					continue
				}
				if updateSaxState(e) {
					updateAudioOutput(saxState)
				}
			}
			SaxStateLayout(gtx, saxState)
			// register for keyboard input
			eventArea := clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Push(ops)
			key.FocusOp{Tag: w}.Add(ops)
			// TODO get the key set from saxKeyMap
			key.InputOp{Tag: w, Keys: key.Set(saxKeySet)}.Add(ops)
			eventArea.Pop()
			e.Frame(gtx.Ops)
		}
	}
}

func updateAudioOutput(s SaxState) {
	// for now: treat the buttons as a binary counter
	freq := playingPitch(s)
	if freq != 0 {
		saxAudioController.notes <- Note{Freq: float32(freq), Vol: 0.1}
	} else {
		saxAudioController.pause <- 1
	}
}

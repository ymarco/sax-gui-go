package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/widget/material"
)

func main() {
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

var maroon = color.NRGBA{R: 127, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 50, G: 50, B: 255, A: 255}

type Title struct {
	color color.NRGBA
}

func NewTitle() Title {
	var res Title
	res.color = maroon
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
				self.color = blue
			case pointer.Release:
				self.color = maroon
			}
		}
	}
	title.Color = self.color
	title.Alignment = text.Middle
	dims := title.Layout(gtx)

	pointer.InputOp{Tag: self, Types: pointer.Press | pointer.Release}.Add(gtx.Ops)
	return dims
}

func run(w *app.Window) error {
	var ops op.Ops
	title := NewTitle()
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			title.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

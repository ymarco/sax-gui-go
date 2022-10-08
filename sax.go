package main

import (
	"math"
)

// this is based on the keys of a Roland Aerophone Mini
type SaxNoteButtons [7]bool // counted from the top; pressing only [0] makes a B4

type SaxState struct {
	// note buttons are the ones that control the notes like a flute
	noteButtons SaxNoteButtons
	// aux buttons are ones like the octave key
	auxButtons [4]bool
	// TODO add a "is blowing" key. Right now there's no key for the left thumb.
	// Maybe we can catch RightAlt.
}

var saxState SaxState

const (
	DownOctaveKey = iota
	UpOctaveKey
	FlatKey
	SharpKey
)

var saxKeyMap = map[string]*bool{
	// left hand
	"U": &saxState.noteButtons[0],
	"E": &saxState.noteButtons[1],
	"O": &saxState.noteButtons[2],
	"A": &saxState.auxButtons[SharpKey],
	";": &saxState.auxButtons[FlatKey],

	"Space": &saxState.auxButtons[UpOctaveKey],
	// TODO find another left thumb key for DownOctaveKey
	"H": &saxState.noteButtons[3],
	"T": &saxState.noteButtons[4],
	"N": &saxState.noteButtons[5],
	"S": &saxState.noteButtons[6],
}

const saxKeySet = "U|E|O|A|;|Space|H|T|N|S"

type ButtonDrawingInstruction struct {
	valPtr *bool
	size   float64
}

// this is a constant
var saxButtonDrawingInstructions = []ButtonDrawingInstruction{
	{valPtr: &saxState.auxButtons[UpOctaveKey], size: 0.5},
	{valPtr: &saxState.noteButtons[0], size: 1.0},
	{valPtr: &saxState.noteButtons[1], size: 1.0},
	{valPtr: &saxState.noteButtons[2], size: 1.0},
	{valPtr: &saxState.auxButtons[FlatKey], size: 0.5},
	{valPtr: &saxState.auxButtons[SharpKey], size: 0.5},
	{valPtr: &saxState.noteButtons[3], size: 1.0},
	{valPtr: &saxState.noteButtons[4], size: 1.0},
	{valPtr: &saxState.noteButtons[5], size: 1.0},
	{valPtr: &saxState.noteButtons[6], size: 1.0},
}

// Thank god equal temperament is easy
func semitoneIntervalFrom(src float64, interval int) float64 {
	return src * math.Pow(2, float64(interval)/12.0)
}

var (
	C5sharp = semitoneIntervalFrom(A4, 4) // highest sax note without mod keys
	C5      = semitoneIntervalFrom(C5sharp, -1)
	B4      = semitoneIntervalFrom(C5, -1)
	A4sharp = semitoneIntervalFrom(A4, 1)
	A4      = 440.0
	G4sharp = semitoneIntervalFrom(A4, -1)
	G4      = semitoneIntervalFrom(A4, -2)
	F4sharp = semitoneIntervalFrom(A4, -3)
	F4      = semitoneIntervalFrom(A4, -4)
	E4      = semitoneIntervalFrom(A4, -5)
	D4Sharp = semitoneIntervalFrom(A4, -6)
	D4      = semitoneIntervalFrom(A4, -7)
	C4sharp = semitoneIntervalFrom(A4, -8)
	C4      = semitoneIntervalFrom(A4, -9)
	B3      = semitoneIntervalFrom(A4, -10)
)
var saxFingeringsMap = map[SaxNoteButtons]float64{
	// {false, false, false, false, false, false, false}: C5sharp,
	{false, false, false, false, false, false, false}: 0.0,
	{false, true, false, false, false, false, false}:  C5,
	{true, false, false, false, false, false, false}:  B4,
	{true, true, false, false, false, false, false}:   A4,
	{true, true, true, false, false, false, false}:    G4,
	{true, true, true, true, false, false, false}:     F4,
	{true, true, true, true, true, false, false}:      E4,
	{true, true, true, true, true, true, false}:       D4,
	{true, true, true, true, true, true, true}:        C4,
}

// Return the pitch that the sax is playing based on s.
// A return pitch of 0 means it's not playing anything.
func playingPitch(s SaxState) float64 {
	basePitch, ok := saxFingeringsMap[s.noteButtons]
	if !ok {
		basePitch = C5sharp // TODO this might not be correct
	}
	modifier := 1.0
	if s.auxButtons[FlatKey] {
		modifier = semitoneIntervalFrom(modifier, -1)
	}
	if s.auxButtons[SharpKey] {
		modifier = semitoneIntervalFrom(modifier, 1)
	}
	if s.auxButtons[DownOctaveKey] {
		modifier = semitoneIntervalFrom(modifier, -12)
	}
	if s.auxButtons[UpOctaveKey] {
		modifier = semitoneIntervalFrom(modifier, 12)
	}
	return basePitch * modifier
}

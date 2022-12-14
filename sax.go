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
	valPtr  *bool
	size    float64
	xOffset float32
	yOffset float32
}

// this is a constant
var saxButtonDrawingInstructions = []ButtonDrawingInstruction{
	{valPtr: &saxState.auxButtons[UpOctaveKey], size: 40, xOffset: 30, yOffset: 00},
	{valPtr: &saxState.noteButtons[0], size: 60, xOffset: 20, yOffset: 15},
	{valPtr: &saxState.noteButtons[1], size: 60, xOffset: 20, yOffset: 10},
	{valPtr: &saxState.noteButtons[2], size: 60, xOffset: 50, yOffset: -15},
	{valPtr: &saxState.auxButtons[SharpKey], size: 40, xOffset: 75, yOffset: 10},
	{valPtr: &saxState.auxButtons[FlatKey], size: 40, xOffset: 75, yOffset: 00},
	{valPtr: &saxState.noteButtons[3], size: 60, xOffset: 50, yOffset: 15},
	{valPtr: &saxState.noteButtons[4], size: 60, xOffset: 50, yOffset: 20},
	{valPtr: &saxState.noteButtons[5], size: 60, xOffset: 25, yOffset: 30},
	{valPtr: &saxState.noteButtons[6], size: 60, xOffset: 10, yOffset: 10},
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
	// HACK in practice this wouldn't actually give C5sharp; see playingPitch
	{false, false, false, false, false, false, false}: C5sharp,
	{false, true, false, false, false, false, false}:  C5,
	{true, false, false, false, false, false, false}:  B4,
	{true, true, false, false, false, false, false}:   A4,
	{true, true, true, false, false, false, false}:    G4,
	{true, true, true, true, false, false, false}:     F4,
	// HACK allow playing the low notes when not all the previous buttons are
	// pressed. Standard keyboards detect only up to 6 presses, and we want to
	// allow octave & sharp/flat buttons too.
	//
	// I chose to redact keys before the last until only 4 are pressed.
	{true, true, true, true, true, false, false}:  E4, // original
	// {true, true, true, false, true, false, false}: E4, // with 1 redacted
	{true, true, true, true, true, true, false}:   D4, // original
	// {true, true, true, true, false, true, false}:  D4, // with 1 redacted
	// {true, true, true, false, false, true, false}: D4, // with 2 redacted
	{true, true, true, true, true, true, true}:    C4, // original
	// {true, true, true, true, true, false, true}:   C4, // with 1 redacted
	// {true, true, true, true, false, false, true}:  C4, // with 2 redacted
	// {true, true, true, false, false, false, true}: C4, // with 3 redacted
}

// Return the pitch that the sax should be playing based on s.
// A return pitch of 0 means not playing anything.
func playingPitch(s SaxState) float64 {
	basePitch, ok := saxFingeringsMap[s.noteButtons]
	// HACK make the empty note fingering not play a not at instead of playing
	// C5sharp. The way to play a C5sharp is by playing a fingering with the
	// first button unpressed.
	if s.noteButtons == (SaxNoteButtons{}) {
		basePitch = 0
	}
	if !ok {
		// the fingering has an unrecognized gap. try to see if we get a good
		// fingering by removing the last button
		var noteButtons SaxNoteButtons
		copy(noteButtons[:], s.noteButtons[:])
		for i := len(noteButtons) - 1; i >= 0; i-- {
			noteButtons[i] = false
			if p, ok := saxFingeringsMap[noteButtons]; ok {
				basePitch = p
				break
			}
		}

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

package main

import (
	"github.com/bspaans/bleep/audio"
	"github.com/bspaans/bleep/channels"
	"github.com/bspaans/bleep/controller"
	"github.com/bspaans/bleep/generators"
	"github.com/bspaans/bleep/synth"
	"time"
)

// TODO this is a stand-in for midi. If/when we do a proper synthesizer it'd be
// replaced.
type Note struct {
	Vol  float32
	Freq float32
}

// Plays notes from the channel notes, synchronically (never returns by itself)
// A note with freq or volume 0 means pause playing.
// When a value is recieved from quit, return.
func StreamingPlayer(notes chan Note, quit chan int) {
	// bleep setup copied from their examples/sinewave/sinewave.go
	cfg := audio.NewAudioConfig()
	instr := generators.NewSineWaveOscillator
	channel := channels.NewPolyphonicChannel()
	channel.SetInstrument(instr)
	mixer := synth.NewMixer()
	mixer.AddChannel(channel)
	synth := synth.NewSynth(cfg)
	synth.Mixer = mixer
	synth.EnableSDLSink()
	ctrl := controller.NewController(cfg)
	ctrl.Synth = synth

	// Playing individual notes with mixer.NoteOn/NoteOff produces artifacts
	// when the notes change. So instead we play an A4 note and bend it wherever
	// we like, and only call NoteOff when we stop playing
	noteOn := false

	// And start the synthesizer in the background
	go ctrl.StartSynth()
	for {
		select {
		case note := <-notes:
			if note.Freq == 0 || note.Vol == 0 {
				// stop playing
				mixer.NoteOff(0, 69 /* A4 */)
				noteOn = false
			} else {
				if !noteOn {
					mixer.NoteOn(0, 69 /* A4 */, 1.0)
				}
			}
			mixer.SetPitchbend(0, float64(note.Freq) / A4)
		case <-quit:
			ctrl.Quit()
			return
		}
	}

}

// A controller controls the audio player. See StreamingPlayer for what the
// channels do.
type NoteStreamAudioController struct {
	notes chan Note
	quit  chan int
}

// BufferedStreamAudioPlayer joins together successive notes that come within
// timeout to avoid jarring audio output
func BufferedStreamAudioPlayer(c NoteStreamAudioController, timeout time.Duration) {
	playerNotes := make(chan Note)
	playerQuit := make(chan int)
	var prevNote Note
	go StreamingPlayer(playerNotes, playerQuit)
	for {
		select {
		case <-c.quit:
			playerQuit <- 1
			return
		case note := <-c.notes:
			deadline := time.After(timeout)
		JoinNotes:
			for {
				select {
				case note = <-c.notes:
					continue
				case <-deadline:
					// no other notes
					break JoinNotes
				}
			}
			if note == prevNote {
				continue
			}
			playerNotes <- note
			prevNote = note
		}
	}
}

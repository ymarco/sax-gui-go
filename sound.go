package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/oto/v2"
)

// Stolen from https://github.com/hajimehoshi/oto
//
// FloatBufferTo16BitLE is a naive helper method to convert []float32 buffers to
// 16-bit little-endian, but encoded in byte buffer
//
// Appends the encoded bytes into "to" slice, allowing you to preallocate the
// capacity or just use nil
func FloatBufferTo16BitLE(from []float32, to []byte) []byte {
	for _, v := range from {
		var uv int16
		if v < -1.0 {
			uv = -math.MaxInt16 // we are a bit lazy: -1.0 is encoded as -32767, as this makes math easier, and -32768 is unused
		} else if v > 1.0 {
			uv = math.MaxInt16
		} else {
			uv = int16(v * math.MaxInt16)
		}
		to = append(to, byte(uv&255), byte(uv>>8))
	}
	return to
}

type SineWave struct {
	// Frequency in Hz
	Freq        float32
	// Volume (up to 1.0)
	Vol float32
	SampleRate  int
	SamplesRead int64
}

func (self *SineWave) Read(buf []byte) (int, error) {

	bufFloat := make([]float32, cap(buf)/2)
	// bufFloat [len(buf)/2]float32

	var i int64
	for i = self.SamplesRead; i < self.SamplesRead+int64(cap(bufFloat)); i++ {
		bufFloat[i] = float32(math.Sin(2 * math.Pi * (float64(i) / float64(self.SampleRate)) * float64(self.Freq) * self.Vol))
	}
	FloatBufferTo16BitLE(bufFloat, buf[:0])

	return cap(buf), nil
}

func testSound() {
	// Prepare an Oto context (this will use your default audio device) that will
	// play all our sounds. Its configuration can't be changed later.

	// Usually 44100 or 48000. Other values might cause distortions in Oto
	samplingRate := 44100

	// Number of channels (aka locations) to play sounds from. Either 1 or 2.
	// 1 is mono sound, and 2 is stereo (most speakers are stereo).
	numOfChannels := 1

	// Bytes used by a channel to represent one sample. Either 1 or 2 (usually 2).
	audioBitDepth := 2

	wave := SineWave{Freq: 440, SampleRate: samplingRate}

	otoCtx, readyChan, err := oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-readyChan

	// Create a new 'player' that will handle our sound. Paused by default.
	player := otoCtx.NewPlayer(&wave)

	// Play starts playing the sound and returns without waiting for it (Play() is async).
	player.Play()

	// We can wait for the sound to finish playing using something like this
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
}
// TODO this is a stand-in for midi
type Note struct {
	vol float32
	freq float32
}
func StreamingPlayer(freqs chan Note) {
	for ()
}

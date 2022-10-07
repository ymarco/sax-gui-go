package main

import (
	"log"
	"math"

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
	input    *SineWaveInputGenerator
	smoother Smoother
	// Volume (up to 1.0)
	Vol        float32
	SampleRate int
}

func NewSineWave(freq, vol float32, sampleRate int) SineWave {
	input := &SineWaveInputGenerator{a: SineWaveInputCoeff(freq, sampleRate)}
	return SineWave{Vol: vol, SampleRate: sampleRate,
		input:    input,
		smoother: NewSmoother(input, 1000)}
}

func (self *SineWave) Read(buf []byte) (int, error) {

	bufFloat := make([]float32, cap(buf)/2)
	// bufFloat [len(buf)/2]float32

	var i int64
	for i = 0; i < int64(cap(bufFloat)); i++ {
		bufFloat[i] = float32(math.Sin(
			2*math.Pi*
				float64(self.input.apply()))) * self.Vol
	}
	FloatBufferTo16BitLE(bufFloat, buf[:0])

	return cap(buf), nil
}

type SineWaveInputGenerator struct {
	// the frequency in Hz
	a float32
	// implementation detail to allow smooth transitions
	b           float32
	samplesRead int64
}

func SineWaveInputCoeff(freq float32, sampleRate int) float32 {
	return freq / float32(sampleRate)
}

// TODO add a smoothing function to this
func (self *SineWaveInputGenerator) apply() float32 {
	res := self.a*float32(self.samplesRead) + self.b
	self.samplesRead += 1
	return res
}
func (self *SineWaveInputGenerator) transitionInto(aNew float32) {
	aOld := self.a
	bOld := self.b
	self.a = aNew
	self.b = aOld*float32(self.samplesRead) + bOld - aNew*float32(self.samplesRead)
}
// TODO the whole smoother interface isn't working
type InputGenerator interface {
	apply() float32
}

type Smoother struct {
	src     InputGenerator
	history []float32
	i       int
}

func NewSmoother(src InputGenerator, historyLen int) Smoother {
	s := Smoother{src: src, history: make([]float32, historyLen)}
	// for i := 0; i < historyLen-1; i++ {
	// 	s.history[i] = src.apply()
	// }
	// s.i = historyLen - 1
	return s
}
func (self *Smoother) apply() float32 {
	self.history[self.i] = self.src.apply()
	var sum float32 = 0.0
	for i := 0; i< len(self.history); i++ {
		sum += self.history[(i + 100) % len(self.history)]
		// log.Println(i)
		// sum += self.history[i]
	}
	self.i = (self.i + 1) % len(self.history)
	return sum / float32(len(self.history))
}

// TODO this is a stand-in for midi
type Note struct {
	Vol  float32
	Freq float32
}

func StreamingPlayer(notes chan Note, pause chan int, quit chan int) {
	// Prepare an Oto context (this will use your default audio device) that will
	// play all our sounds. Its configuration can't be changed later.

	// Usually 44100 or 48000. Other values might cause distortions in Oto
	samplingRate := 44100

	// Number of channels (aka locations) to play sounds from. Either 1 or 2.
	// 1 is mono sound, and 2 is stereo (most speakers are stereo).
	numOfChannels := 1

	// Bytes used by a channel to represent one sample. Either 1 or 2 (usually 2).
	audioBitDepth := 2

	otoCtx, readyChan, err := oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-readyChan
	var player oto.Player = nil
	wave := NewSineWave(0, 1.0, samplingRate)
	for {
		select {
		case note := <-notes:
			if player != nil {
				player.Pause()
				// account for unplayed samples
				wave.input.samplesRead -= int64(player.UnplayedBufferSize() / audioBitDepth)
			}
			wave.input.transitionInto(SineWaveInputCoeff(note.Freq, samplingRate))
			// TODO handle note.Vol
			player = otoCtx.NewPlayer(&wave)
			log.Println(wave)
			player.Play() // async, doesn't block
		case <-pause:
			if player != nil {
				player.Pause()
			}
		case <-quit:
			if player != nil {
				player.Pause()
			}
			return
		}
	}

}

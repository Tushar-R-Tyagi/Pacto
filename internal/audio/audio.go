package audio

import (
	"encoding/binary"
	"math"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

var (
	startChime *beep.Buffer
	endChime   *beep.Buffer
)

func Init() error {
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Millisecond*100))

	// Create sounds directory and generate distinct chimes if missing.
	os.MkdirAll("sounds", 0755)
	if _, err := os.Stat("sounds/start_chime.wav"); os.IsNotExist(err) {
		generateChime("sounds/start_chime.wav", []float64{880, 1320}, 0.15, 0.5)
	}
	if _, err := os.Stat("sounds/end_chime.wav"); os.IsNotExist(err) {
		generateChime("sounds/end_chime.wav", []float64{660, 440}, 0.2, 0.6)
	}

	startFile, err := os.Open("sounds/start_chime.wav")
	if err != nil {
		return err
	}
	defer startFile.Close()
	startStreamer, format, err := wav.Decode(startFile)
	if err != nil {
		return err
	}
	startChime = beep.NewBuffer(format)
	startChime.Append(startStreamer)

	endFile, err := os.Open("sounds/end_chime.wav")
	if err != nil {
		return err
	}
	defer endFile.Close()
	endStreamer, _, err := wav.Decode(endFile)
	if err != nil {
		return err
	}
	endChime = beep.NewBuffer(format)
	endChime.Append(endStreamer)

	return nil
}

func generateChime(path string, freqs []float64, duration, volume float64) {
	sampleRate := 44100
	parts := len(freqs)
	if parts == 0 {
		return
	}
	samplesPerPart := int(float64(sampleRate) * duration)
	numSamples := samplesPerPart * parts

	f, _ := os.Create(path)
	defer f.Close()

	dataSize := numSamples * 2
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	f.Write([]byte("WAVE"))
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*2))
	binary.Write(f, binary.LittleEndian, uint16(2))
	binary.Write(f, binary.LittleEndian, uint16(16))
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))

	for partIndex, freq := range freqs {
		for i := 0; i < samplesPerPart; i++ {
			globalIndex := partIndex*samplesPerPart + i
			t := float64(i) / float64(sampleRate)
			fade := 1.0 - (float64(globalIndex) / float64(numSamples))
			sample := int16(volume * fade * 32767.0 * math.Sin(2.0*math.Pi*freq*t))
			binary.Write(f, binary.LittleEndian, sample)
		}
	}
}

func PlayStartChime() {
	if startChime == nil {
		return
	}
	streamer := startChime.Streamer(0, startChime.Len())
	speaker.Play(streamer)
}

func PlayEndChime() {
	if endChime == nil {
		return
	}
	streamer := endChime.Streamer(0, endChime.Len())
	speaker.Play(streamer)
}

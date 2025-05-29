package audioproc

import (
	"fmt"
	"math"

	"github.com/digital-dream-labs/opus-go/opus"
	"github.com/maxhawkins/go-webrtcvad"
)

type AudioProcessor struct {
	// this vad is for audio processing, NOT for detecting end of speech
	vad        *webrtcvad.VAD
	opusStream *opus.OggStream

	pastFirstChunk bool
	isOpus         bool

	runningRMS         float64
	targetRMS          float64
	highPassFilter     *biquadFilter
	preHighPassFilter  *biquadFilter
	smoothingAlpha     float64
	maxGainChange      float64
	noiseGateThreshold float64
	noiseGateReduction float64
	hpAlpha            float64
	prevIn             float64
	currentGain        float64
	prevOut            float64
}

func NewAudioProcessor(sampleRate int, cutoffHz float64, vadMode int) (*AudioProcessor, error) {
	k := math.Tan(math.Pi * cutoffHz / float64(sampleRate))
	alpha := k / (1 + k)

	v, err := webrtcvad.New()
	if err != nil {
		return nil, err
	}
	if err := v.SetMode(vadMode); err != nil {
		return nil, err
	}

	return &AudioProcessor{
		vad:                v,
		opusStream:         &opus.OggStream{},
		runningRMS:         1,
		targetRMS:          2000,
		smoothingAlpha:     0.1,
		maxGainChange:      2,
		noiseGateThreshold: 500,
		noiseGateReduction: 0.5,
		hpAlpha:            alpha,
		currentGain:        1.0, // start at 1
		highPassFilter:     newBiquadHighPass(float64(sampleRate), cutoffHz),
		preHighPassFilter:  newBiquadHighPass(float64(sampleRate), cutoffHz),
	}, nil
}

func (rawr *AudioProcessor) ProcessAudio(buf []byte) []byte {
	if !rawr.pastFirstChunk {
		rawr.isOpus = OpusDetect(buf)
	}
	if rawr.isOpus {
		var err error
		buf, err = rawr.opusStream.Decode(buf)
		if err != nil {
			fmt.Println("opus stream decode error (ProcessAudio):", err)
		}
	}
	frames := SplitIntoFrames(buf, 320)
	var output []byte
	for _, frame := range frames {
		active, err := rawr.vad.Process(16000, frame)
		if err != nil {
			fmt.Println("webrtcvad Process() error (ProcessAudio):", err)
		}
		int16Data := bytesToInt16(frame)
		processed := rawr.processInt16Chunk(int16Data, active)
		output = append(output, int16ToBytes(processed)...)
	}
	rawr.pastFirstChunk = true
	return output
}

// do HP filter, RMS-based normalization, noise gate...
func (cat *AudioProcessor) processInt16Chunk(samples []int16, active bool) []int16 {
	out := make([]int16, len(samples))
	samples = cat.highPassFilter.process(samples)
	rms := computeRMS(samples)

	if active {
		cat.runningRMS = cat.smoothingAlpha*rms + (1-cat.smoothingAlpha)*cat.runningRMS
	}

	var desiredGain float64 = 1.0

	if cat.runningRMS > 0 && cat.runningRMS < cat.targetRMS {
		desiredGain = math.Sqrt(cat.targetRMS / cat.runningRMS)
	}

	// ensure we never reduce gain below 1.0
	if desiredGain < 1.0 {
		desiredGain = 1.0
	}

	// clamp the gain change so it can't jump more than maxGainChange times the old gain
	if desiredGain > cat.currentGain*cat.maxGainChange {
		desiredGain = cat.currentGain * cat.maxGainChange
	}
	cat.currentGain = desiredGain

	// apply high-pass
	samples = cat.highPassFilter.process(samples)

	// scale samples
	for i, s := range samples {
		scaled := float64(s) * cat.currentGain
		// clamp
		if scaled > math.MaxInt16 {
			scaled = math.MaxInt16
		} else if scaled < math.MinInt16 {
			scaled = math.MinInt16
		}
		out[i] = int16(scaled)
	}

	return out
}

func OpusDetect(firstChunk []byte) bool {
	var isOpus bool
	if len(firstChunk) > 0 {
		if firstChunk[0] == 0x4f {
			isOpus = true
		} else {
			isOpus = false
		}
	}
	return isOpus
}

func SplitIntoFrames(buf []byte, frameSize int) [][]byte {
	var frames [][]byte
	for len(buf) >= frameSize {
		frames = append(frames, buf[:frameSize])
		buf = buf[frameSize:]
	}
	return frames
}

func computeRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	var count int // count valid samples
	for _, s := range samples {
		if s > math.MaxInt16-100 || s < math.MinInt16+100 {
			continue // skip clipped samples
		}
		fs := float64(s)
		sum += fs * fs
		count++
	}
	if count == 0 {
		return 0 // avoid division by zero
	}
	return math.Sqrt(sum / float64(count))
}

func bytesToInt16(b []byte) []int16 {
	n := len(b) / 2
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		out[i] = int16(b[2*i]) | int16(b[2*i+1])<<8
	}
	return out
}

func int16ToBytes(samples []int16) []byte {
	out := make([]byte, len(samples)*2)
	for i, s := range samples {
		out[2*i] = byte(s & 0xff)
		out[2*i+1] = byte((s >> 8) & 0xff)
	}
	return out
}

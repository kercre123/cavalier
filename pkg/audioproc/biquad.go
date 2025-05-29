package audioproc

import "math"

type biquadFilter struct {
	a0, a1, a2 float64
	b1, b2     float64
	z1, z2     float64 // filter states
}

func newBiquadHighPass(sampleRate, cutoffHz float64) *biquadFilter {
	omega := 2.0 * math.Pi * cutoffHz / sampleRate
	cosOmega := math.Cos(omega)
	sinOmega := math.Sin(omega)
	alpha := sinOmega / (2.0 * 0.707)

	a0 := 1.0 + alpha
	a1 := -2.0 * cosOmega
	a2 := 1.0 - alpha
	b0 := (1.0 + cosOmega) / 2.0
	b1 := -(1.0 + cosOmega)
	b2 := (1.0 + cosOmega) / 2.0

	// normalize coefficients
	return &biquadFilter{
		a0: b0 / a0,
		a1: b1 / a0,
		a2: b2 / a0,
		b1: a1 / a0,
		b2: a2 / a0,
		z1: 0.0,
		z2: 0.0,
	}
}

func (f *biquadFilter) process(samples []int16) []int16 {
	out := make([]int16, len(samples))
	for i, sample := range samples {
		x := float64(sample)

		// biquad filter difference equation
		y := f.a0*x + f.z1
		f.z1 = f.a1*x + f.z2 - f.b1*y
		f.z2 = f.a2*x - f.b2*y

		// clamp to int16 range
		if y > math.MaxInt16 {
			y = math.MaxInt16
		} else if y < math.MinInt16 {
			y = math.MinInt16
		}
		out[i] = int16(y)
	}
	return out
}

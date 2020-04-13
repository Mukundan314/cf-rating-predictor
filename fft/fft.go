package fft

import (
	"math"
	"math/cmplx"
)

func FFT(a []complex128, inv bool) []complex128 {
	n := len(a)

	b := make([]complex128, n)
	copy(b, a)

	w := make([]complex128, n>>1)
	for i := 0; i < n>>1; i++ {
		if inv {
			w[i] = cmplx.Rect(1, -2.0*math.Pi*float64(i)/float64(n))
		} else {
			w[i] = cmplx.Rect(1, 2.0*math.Pi*float64(i)/float64(n))
		}
	}

	rev := make([]int, n)
	for i := 0; i < n; i++ {
		rev[i] = rev[i>>1] >> 1
		if i&1 == 1 {
			rev[i] |= n >> 1
		}
		if i < rev[i] {
			b[i], b[rev[i]] = b[rev[i]], b[i]
		}
	}

	for step := 2; step <= n; step <<= 1 {
		half, diff := step>>1, n/step
		for i := 0; i < n; i += step {
			pw := 0
			for j := i; j < i+half; j++ {
				v := b[j+half] * w[pw]
				b[j+half] = b[j] - v
				b[j] += v
				pw += diff
			}
		}
	}

	if inv {
		for i := 0; i < n; i++ {
			b[i] /= complex(float64(n), 0.0)
		}
	}

	return b
}

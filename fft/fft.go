package fft

import (
	"math"
	"math/cmplx"
)

func FFT(a []complex128, inv bool) {
	n := len(a)

	w := make([]complex128, n>>1)
	for i := 0; i < n>>1; i++ {
		if inv {
			w[i] = cmplx.Rect(1, (-2.0*math.Pi*float64(i))/float64(n))
		} else {
			w[i] = cmplx.Rect(1, (2.0*math.Pi*float64(i))/float64(n))
		}
	}

	rev := make([]int, n)
	for i := 0; i < n; i++ {
		rev[i] = rev[i>>1] >> 1
		if i&1 == 1 {
			rev[i] |= n >> 1
		}
		if i < rev[i] {
			a[i], a[rev[i]] = a[rev[i]], a[i]
		}
	}

	for step := 2; step <= n; {
		half, diff := step>>1, n/step
		for i := 0; i < n; i += step {
			pw := 0
			for j := i; j < i+half; j++ {
				v := a[j+half] * w[pw]
				a[j+half] = a[j] - v
				a[j] += v
				pw += diff
			}
		}
		step <<= 1
	}

	if inv {
		for i := 0; i < n; i++ {
			a[i] /= complex(float64(n), 0.0)
		}
	}
}

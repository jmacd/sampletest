package sampletest

import (
	"fmt"
	"math/bits"
	"math/rand"
	"sort"
	"testing"

	"github.com/jmacd/sampletest/kolmogorov"
	"gonum.org/v1/gonum/stat/distuv"
)

// K-S test

const (
	repeats  = 10
	trials   = 100
	expected = 100
)

type uniformProbTester struct {
	prob float64
	rnd  *rand.Rand
}

type trialResults []int

func computeResults(prob float64, f func() bool) (r []trialResults) {
	tosses := int(expected / prob)

	for j := 0; j < repeats; j++ {
		var results []int
		for i := 0; i < trials; i++ {
			successes := 0
			for j := 0; j < tosses; j++ {
				if f() {
					successes++
				}
			}
			results = append(results, successes)
		}

		sort.Ints(results)
		r = append(r, trialResults(results))
	}
	return
}

func newUniformProbTester(src rand.Source, prob float64) *uniformProbTester {
	return &uniformProbTester{
		prob: prob,
		rnd:  rand.New(src),
	}
}

func (u *uniformProbTester) decision() bool {
	return u.rnd.Float64() < u.prob
}

func newSource() rand.Source {
	return rand.NewSource(77777677777)
}

func TestSampling(t *testing.T) {
	for nlp := 0; nlp < 10; nlp++ {
		testSampling(t, nlp)
	}
}

func testSampling(t *testing.T, neglogprob int) {
	prob := 1.0 / float64(uint64(1)<<neglogprob)

	fmt.Printf("At probability 1-in-2**%d\n", neglogprob)
	func() {
		res := computeResults(prob, newUniformProbTester(newSource(), prob).decision)

		var dvals []float64
		for i := 0; i < repeats; i++ {
			dvals = append(dvals, tester(prob, res[i]))
		}
		retester("math.Float64", dvals)
	}()

	run := func(name string, f func() bool) {
		res := computeResults(prob, f)
		var dvals []float64
		for i := 0; i < repeats; i++ {
			dvals = append(dvals, tester(prob, res[i]))
		}
		retester(name, dvals)
	}

	func() {
		src := rand.New(newSource())
		run("LeadingZeros", func() bool {
			return bits.LeadingZeros64(uint64(src.Int63()<<1)) >= neglogprob
		})
	}()

	func() {
		src := rand.New(newSource())
		run("TrailingZeros", func() bool {
			return bits.TrailingZeros64(uint64(src.Int63())) >= neglogprob
		})
	}()

	func() {
		src := rand.New(newSource())
		run("LeadingOnes", func() bool {
			return bits.LeadingZeros64(0^uint64(src.Int63()<<1)) >= neglogprob
		})
	}()

	func() {
		src := rand.New(newSource())
		run("TrailingOnes", func() bool {
			return bits.TrailingZeros64(0^uint64(src.Int63())) >= neglogprob
		})
	}()

	func() {
		src := rand.New(newSource())
		avail := 0
		state := int64(0)
		run("ReusedBits", func() bool {
			r := 0
			for {
				if avail == 0 {
					avail = 63
					state = src.Int63()
				}
				one := state&1 == 1
				state >>= 1
				avail--
				if one {
					break
				}
				r++
			}
			return r >= neglogprob
		})
	}()

	func() {
		cnt := 0
		run("NotRandom", func() bool {
			cnt++
			cnt = cnt % (1 << neglogprob)
			if cnt == 0 {
				return true
			}
			return false
		})
	}()

	func() {
		src := rand.New(newSource())
		run("BiasedTrue", func() bool {
			if src.Float64() < 0.01 {
				return true
			}
			return bits.LeadingZeros64(uint64(src.Int63()<<1)) >= neglogprob
		})
	}()

	func() {
		src := rand.New(newSource())
		run("BiasedFalse", func() bool {
			if src.Float64() < 0.1 {
				return false
			}
			return bits.LeadingZeros64(uint64(src.Int63()<<1)) >= neglogprob
		})
	}()
}

func tester(prob float64, results trialResults) float64 {
	tosses := expected / prob

	dist := distuv.Binomial{
		N: tosses,
		P: prob,
	}

	// Like Knuth 3.3.1 algorithm B, one-sample, but without the
	// sqrt(trials) term, thus using the exact Kolmogorov
	// distribution instead of K+ and K- like Knuth.
	d := 0.0

	for i := 0; i < trials; i++ {
		val := results[i]
		for i < trials-1 && results[i+1] == val {
			i++ // Scanning past duplicates
		}

		if dPlus := (float64(i+1) / trials) - dist.CDF(float64(results[i])); dPlus > d {
			d = dPlus
		}
		if dMinus := dist.CDF(float64(results[i])) - (float64(i) / trials); dMinus > d {
			d = dMinus
		}
	}

	return d
}

func retester(name string, results []float64) float64 {
	d := 0.0

	for i := 0; i < repeats; i++ {
		val := results[i]
		for i < repeats-1 && results[i+1] == val {
			i++ // Scanning past duplicates
		}

		if dPlus := (float64(i+1) / repeats) - kolmogorov.K(trials, results[i]); dPlus > d {
			d = dPlus
		}
		if dMinus := kolmogorov.K(trials, results[i]) - (float64(i) / repeats); dMinus > d {
			d = dMinus
		}
	}

	fmt.Printf("%s: K-S multi D %f%%\n", name, 100*kolmogorov.K(repeats, d))
	return d
}

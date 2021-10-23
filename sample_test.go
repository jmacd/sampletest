package sampletest

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/jmacd/sampletest/kolmogorov"
	"gonum.org/v1/gonum/stat/distuv"
)

// K-S test

const (
	repeats = 10
	trials  = 1000
	tosses  = 1000

	neglogprob = 3
)

type uniformProbTester struct {
	prob float64
	rnd  *rand.Rand
}

type trialResults []int

func computeResults(f func() bool) (r []trialResults) {
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
	prob := 1.0 / float64(uint64(1)<<neglogprob)

	fmt.Println("trials", trials, "prob", prob)

	// sixtyThree := func(rnd *rand.Rand) (x int64) {
	// 	for i := 0; i < 63; i++ {
	// 		if rnd.Float64() < 0.5 {
	// 			x |= int64(1) << i
	// 		}
	// 	}
	// 	return
	// }
	// _ = sixtyThree

	res := computeResults(newUniformProbTester(newSource(), prob).decision)

	_ = kolmogorov.K

	for i := 0; i < repeats; i++ {
		tester("math.Float64", prob, res[i])
	}

	// tester("bits.ExpensiveLeadingZeros64", prob, func() bool {
	// 	return bits.LeadingZeros64(uint64(sixtyThree())<<1) >= neglogprob
	// })

	// tester("bits.LeadingZeros64", prob, func() bool {
	// 	return bits.LeadingZeros64(uint64(rand.Int63())<<1) >= neglogprob
	// })

	// tester("bits.ExpensiveTrailingZeros64", prob, func() bool {
	// 	return bits.TrailingZeros64(uint64(sixtyThree())) >= neglogprob
	// })

	// tester("bits.TrailingZeros64", prob, func() bool {
	// 	return bits.TrailingZeros64(uint64(rand.Int63())) >= neglogprob
	// })

	// tester("bits.ExpensiveLeadingOnes64", prob, func() bool {
	// 	return bits.LeadingZeros64(0^uint64(sixtyThree())<<1) >= neglogprob
	// })

	// tester("bits.LeadingOnes64", prob, func() bool {
	// 	return bits.LeadingZeros64(0^uint64(rand.Int63())<<1) >= neglogprob
	// })

	// tester("bits.ExpensiveTrailingOnes64", prob, func() bool {
	// 	return bits.TrailingZeros64(0^uint64(sixtyThree())) >= neglogprob
	// })

	// tester("bits.TrailingOnes64", prob, func() bool {
	// 	return bits.TrailingZeros64(0^uint64(rand.Int63())) >= neglogprob
	// })
}

func tester(name string, prob float64, results trialResults) (float64, float64) {

	dist := distuv.Binomial{
		N: tosses,
		P: prob,
	}
	sqrtTrials := math.Sqrt(trials)

	// From Knuth 3.3.1 algorithm B, one-sample.
	kPlus := 0.0
	kMinus := 0.0

	for i := 0; i < trials; i++ {
		val := results[i]
		for i < trials-1 && results[i+1] == val {
			i++ // Scanning past duplicates
		}

		if dPlus := (float64(i+1) / trials) - dist.CDF(float64(results[i])); dPlus > kPlus {
			kPlus = dPlus
		}
		if dMinus := dist.CDF(float64(results[i])) - (float64(i) / trials); dMinus > kMinus {
			kMinus = dMinus
		}
	}

	kPlus *= sqrtTrials
	kMinus *= sqrtTrials

	kProb := func(s float64) float64 {
		// Knuth 3.3.1 equation 27
		return 1 - math.Exp(-2*s*s)*(1-(2*s)/(3*sqrtTrials))
	}

	maxK := kPlus
	if maxK < kMinus {
		maxK = kMinus
	}

	fmt.Println(name, "K-S single K+", kPlus, kProb(kPlus))
	fmt.Println(name, "K-S single K-", kMinus, kProb(kMinus))
	return kPlus, kMinus
}

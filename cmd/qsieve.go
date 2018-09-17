package main

import (
    "github.com/yabberyabber/factoring"
    "math"
    "fmt"
    "time"
    "sync/atomic"
)

const (
    smoothnessCap = 63
    // toFactor = 90283
    toFactor = 502560280658509
)

var done atomic.Value

type divisionTrial struct {
    divisor uint
    ack     uint
    vect factoring.PrimeVector
}

type factorPair struct {
    a, b uint
}

func main() {
    // Initialize the naieve divisor with the first SMOOTHNESS primes
    factoring.GetNthPrime(smoothnessCap)

    done.Store(false)

    /*
    for i := uint(1); i < smoothnessCap; i++ {
        fmt.Printf("prime[%d] = %d\n", i, factoring.GetNthPrime(i))
    }
    */

    smoothDivisorsChan := make(chan divisionTrial, 10)
    go GetSmoothDivisors(smoothDivisorsChan)

    filteredSmoothDivisorsChan := make(chan divisionTrial, 3)
    go FilterSmoothDivisors(smoothDivisorsChan, filteredSmoothDivisorsChan)

    squareChan := make(chan []divisionTrial, 1)
    go SquareSubsetFilter(filteredSmoothDivisorsChan, squareChan)

    resultChan := make(chan factorPair, 1)
    go DetermineFactor(squareChan, resultChan)

    filteredResultChan := make(chan factorPair, 1)
    go FilterTrivialResults(resultChan, filteredResultChan)

    result := <-filteredResultChan
    fmt.Printf("%d * %d = %d\n", result.a, result.b,
               result.a * result.b)

    done.Store(true)
}

func GetSmoothDivisors(smoothResults chan divisionTrial) {
    candidate := uint(math.Sqrt(float64(toFactor))) + 1
    searchStartTime := time.Now()

    for {
        if done.Load().(bool) {
            return
        }

        result := (candidate * candidate) - toFactor

        vect, err := factoring.FactorSmoothNumber(result, smoothnessCap)
        if err == nil {
            newEntry := divisionTrial{
                divisor: candidate,
                vect:  vect,
                ack:     result,
            }

            fmt.Printf("Found new smooth factor (took %v)\n",
                       time.Since(searchStartTime))
            smoothResults <-newEntry
            searchStartTime = time.Now()
        }

        candidate += 1
    }
}

// fund a subset of |smooth| that results in a square.  for now use brute force
func SquareSubsetFilter(smooth chan divisionTrial, squares chan []divisionTrial) {
    entries := []divisionTrial{}

    for {
        if done.Load().(bool) {
            return
        }

        newEntry := <-smooth
        entries = append(entries, newEntry)
        // fmt.Printf("Just read from chan %+v\n", newEntry)

        startTime := time.Now()

        for i := uint64(math.Pow(2, float64(len(entries) - 1)));
            i < uint64(math.Pow(2, float64(len(entries))));
            i++ {
            result := factoring.NewPrimeVector()
            for j := uint64(0); j < uint64(len(entries)); j++ {
                if i & (uint64(1) << j) != 0 {
                    result = factoring.JoinVects(result, entries[j].vect)
                }
            }

            // fmt.Printf("%8b: %+v\n", i, result)

            if result.Square() {
                squares <- BuildSubset(entries, i)
            }
        }

        fmt.Printf("No square subset given %d smooth divisions (took %v)\n",
                   len(entries), time.Since(startTime))

    }
}

// Build a subset of the given list using |bitmap| to select elements
func BuildSubset(smooth []divisionTrial, bitmap uint64) []divisionTrial {
    result := []divisionTrial{}

    for i := uint64(0); i < uint64(len(smooth)); i++ {
        if bitmap & (1 << i) != 0 {
            result = append(result, smooth[i])
        }
    }

    return result
}

func DetermineFactor(smoothChan chan []divisionTrial, outChan chan factorPair) {
    for {
        smooth := <-smoothChan

        inProduct := uint(1)
        outProductSquare := uint(1)

        for _, entry := range(smooth) {
            fmt.Printf("%d\n", entry.divisor)
            inProduct *= entry.divisor
            outProductSquare *= entry.ack
        }

        outProduct := uint(math.Sqrt(float64(outProductSquare)))

         a, b := inProduct - outProduct, inProduct + outProduct
         a = GCD(toFactor, a)
         b = GCD(toFactor, b)

         outChan <-factorPair{a: a, b: b}
    }
}

func FilterSmoothDivisors(smoothDivisorsChan,
    filteredSmoothDivisorsChan chan divisionTrial) {
    seen := map[int64]bool{}

    seenCount := 0
    filteredCount := 0

    for trial := range smoothDivisorsChan {
        key := trial.vect.Hash()
        seenCount++
        if !seen[key] {
            filteredSmoothDivisorsChan <-trial
            seen[key] = true
        } else {
            filteredCount++
            fmt.Printf("\n***\nRedundant division trial fitered out (%d of %d)\n***\n",
                       filteredCount, seenCount)
        }
    }
}

func FilterTrivialResults(inChan chan factorPair, outChan chan factorPair) {
    for result := range inChan {
        if result.a != 1 {
            outChan <-result
        } else {
            fmt.Printf("\n***\nTrivial result fitered out\n***\n")
        }
    }
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

func GCD(a, b uint) uint {
    for b != 0 {
        a, b = b, a%b
    }

    return a
}

package main

import (
    "github.com/yabberyabber/factoring"
    "math"
    "fmt"
    "time"
    "sync/atomic"
    "runtime"
)

const (
    smoothnessCap = 42
    // toFactor = 659 * 137 // 90283
    // toFactor = 1249 * 1451
    toFactor = 8009 * 15373
    // toFactor = 102199 * 15373
    // toFactor = 502560280658509

)

var done atomic.Value

type divisionTrial struct {
    divisor uint64
    ack     uint64
    vect factoring.PrimeVector
}

type factorPair struct {
    a, b uint64
}

func main() {
    runtime.GOMAXPROCS(16)

    defer timeTrack(time.Now(), "Factoring")
    // Initialize the naieve divisor with the first SMOOTHNESS primes
    factoring.GetNthPrime(smoothnessCap)

    done.Store(false)

    smoothDivisorsChan := make(chan divisionTrial, 32)
    go GetSmoothDivisors(smoothDivisorsChan)

    filteredSmoothDivisorsChan := make(chan divisionTrial, 32)
    go FilterSmoothDivisors(smoothDivisorsChan, filteredSmoothDivisorsChan)

    accumulatedDivisorsChan := make(chan []divisionTrial, 32)
    go SmoothDivisorsAccumulator(filteredSmoothDivisorsChan, accumulatedDivisorsChan)

    squareChan := make(chan []divisionTrial, 16)
    for i := 0; i < 4; i++ {
        go ComputeSquareSubset(accumulatedDivisorsChan, squareChan)
    }

    resultChan := make(chan factorPair, 4)
    go DetermineFactor(squareChan, resultChan)

    filteredResultChan := make(chan factorPair, 4)
    go FilterTrivialResults(resultChan, filteredResultChan)

    tick := make(chan bool)
    go func() {
        for {
            time.Sleep(10 * time.Second)
            tick <- true
        }
    }()

    OuterLoop:
    for {
        select {
        case <-tick:
            fmt.Printf("Queue of sets to find squares in: %d\n", len(accumulatedDivisorsChan))
        case result := <-filteredResultChan:
            fmt.Printf("%d * %d = %d\n", result.a, result.b,
                       result.a * result.b)
            break OuterLoop
        }
    }

    done.Store(true)
}

func GetSmoothDivisors(smoothResults chan divisionTrial) {
    candidate := uint64(math.Sqrt(float64(toFactor))) + 1
    // searchStartTime := time.Now()

    for i := 0; ; i++{
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

            /*
            fmt.Printf("Found new smooth factor (took %v) (queue %d)\n",
                       time.Since(searchStartTime), len(smoothResults))
                       */
            smoothResults <-newEntry
            // searchStartTime = time.Now()
        }

        candidate += 1
    }
}

func SmoothDivisorsAccumulator(smooth chan divisionTrial, sets chan []divisionTrial) {
    entries := []divisionTrial{}

    for {
        if done.Load().(bool) {
            return
        }

        newEntry := <-smooth
        entries = append(entries, newEntry)

        sets <-entries
        /*
        fmt.Printf("Accumulated divisors (queue %d)\n",
                   len(sets))
                   */
    }
}

// Do the actual computation finding a square subset of |smooth|
func ComputeSquareSubset(entriesChan chan []divisionTrial,
        squares chan []divisionTrial) {
    for {
        if done.Load().(bool) {
            return
        }

        entries := <-entriesChan

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

        fmt.Printf("No square subset given %d smooth divisions (took %v) (queue %d)\n",
                   len(entries), time.Since(startTime), len(entriesChan))
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

        inProduct := uint64(1)
        outProductSquare := uint64(1)

        for _, entry := range(smooth) {
            inProduct *= entry.divisor
            outProductSquare *= entry.ack
        }

        outProduct := uint64(math.Sqrt(float64(outProductSquare)))

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
            /*
            fmt.Printf("Filtered smooth factor (queue %d)\n",
                       len(filteredSmoothDivisorsChan))
                       */
        } else {
            filteredCount++
            fmt.Printf("\n***\nRedundant division trial fitered out (%d of %d)\n***\n",
                       filteredCount, seenCount)
        }
    }
}

func FilterTrivialResults(inChan chan factorPair, outChan chan factorPair) {
    for result := range inChan {
        if result.a != 1 && result.b != 1 {
            outChan <-result
        } else {
            // fmt.Printf("\n***\nTrivial result fitered out\n***\n")
        }
    }
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

func GCD(a, b uint64) uint64 {
    for b != 0 {
        a, b = b, a%b
    }

    return a
}

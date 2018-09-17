package factoring

import "errors"

// A composit number can be represented as the product of many primes. 
type PrimeVector struct {
    factors int64
}

func (p *PrimeVector) getFactor(n uint64) bool {
    return (p.factors & (1 << (n - 1))) != 0
}

func (p *PrimeVector) addFactor(n uint64) {
    p.factors |= 1 << (n - 1)
}

func (p *PrimeVector) Square() bool {
    return p.factors == 0
}

func NewPrimeVector() PrimeVector {
    return PrimeVector{factors: 0}
}

func JoinVects(a, b PrimeVector) PrimeVector {
    return PrimeVector{
        factors: a.factors ^ b.factors,
    }
}

func (p *PrimeVector) Hash() int64 {
    return p.factors
}

func FactorSmoothNumber(n uint64, smoothnessCap uint64) (PrimeVector, error) {
    result := NewPrimeVector()

    for i := uint64(1); i < smoothnessCap && n > 0; i++ {
        nthPrime := GetNthPrime(i)
        for n % nthPrime == 0 {
            n = n / nthPrime
            result.factors ^= (1 << (i - 1))
        }
    }

    if n != 1 {
        return result, errors.New("n is not smooth")
    }

    return result, nil
}


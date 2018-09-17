package factoring

import "math"

// list of the first couple prime numbers
var factors = []uint{
    2, 3, 5, 7,
}

// Get the |n|th prime number using the global cache as much as possible
func GetNthPrime(n uint) uint {
    for uint(len(factors)) < n {
        next_prime_candidate := uint(factors[len(factors) - 1])

        for {
            next_prime_candidate += 2
            if isPrime(next_prime_candidate) {
                factors = append(factors, next_prime_candidate)
                break
            }
        }
    }

    return factors[n - 1]
}

// Check if the given number is prime using the global cache as much as
// possible
func isPrime(n uint) bool {
    if n <= 2 {
        return true
    }

    i := uint(1)
    largest_possible_factor := uint(math.Sqrt(float64(n))) + 1
    for {
        factor := GetNthPrime(i)
        if factor > largest_possible_factor {
            return true
        }

        if n % factor == 0 {
            return false
        }
        i += 1
    }
}

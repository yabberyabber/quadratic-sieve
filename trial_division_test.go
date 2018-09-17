package factoring

import "testing"

func TestIsPrime(t *testing.T) {
    tt := []struct{
        num uint
        primeness bool
    }{
        {
            1, true,
        },
        {
            2, true,
        },
        {
            3, true,
        },
        {
            4, false,
        },
        {
            6, false,
        },
        {
            9, false,
        },
        {
            11, true,
        },
        {
            100, false,
        },
        {
            199, true,
        },
    }

    for _, tc := range tt {
        if isPrime(tc.num) != tc.primeness {
            t.Errorf("isPrime(%d) was %d... expected %d\n",
                     tc.num, isPrime(tc.num), tc.primeness)
        }
    }
}

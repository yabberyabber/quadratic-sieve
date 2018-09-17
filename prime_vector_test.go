package factoring

import "testing"

func TestJoinVects(t *testing.T) {
    vectA := PrimeVector{ factors: 0x02 }
    vectB := PrimeVector{ factors: 0x03 }
    expected := PrimeVector{ factors: 0x01 }

    if JoinVects(vectA, vectB) != expected {
        t.Errorf("Joining didn't do it right!")
    }
}

func TestPrimeVectorBitwiseArithmetic(t *testing.T) {
    vect := PrimeVector{ factors: 0x0002 }

    if vect.getFactor(1) {
        t.Errorf("Didn't expect factor id 1")
    }
    if !vect.getFactor(2) {
        t.Errorf("Expected factor id 2")
    }
    if vect.getFactor(3) {
        t.Errorf("Didn't expect factor id 3")
    }
}

func TestRoughFactor(t *testing.T) {
    tt := []struct {
        input uint
        smooth bool
    }{
        { 6, true, },
        { 9, true, },
        { 502560280658509, false, },
    }

    for _, tc := range tt {
        _, err := FactorSmoothNumber(tc.input, 64)
        if (err == nil) != tc.smooth {
            t.Errorf("Expected smoothness %d but got %d\n",
                     tc.smooth, (err == nil))
        }
    }
}

func TestSmoothFactor(t *testing.T) {
    tt := []struct {
        input uint
        expected int64
    }{
        {
            6,
            0x03,
        },
        {
            36,
            0x00,
        },
        {
            1,
            0x0,
        },
        {
            5,
            0x4,
        },
    }

    for _, tc := range tt {
        vect, err := FactorSmoothNumber(tc.input, 64)
        if err != nil {
            t.Errorf("Was not expecting error: %v", err)
        }
        if vect.factors != tc.expected {
            t.Errorf("Expected %x got %x", tc.expected, vect.factors)
        }
    }
}

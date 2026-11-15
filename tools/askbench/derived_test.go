package main

import "testing"

func TestAnswerNumbersAreOneStepOfTwoCardNumbers(t *testing.T) {
	t.Parallel()
	// The two value cards of the seed's November 2025 and the net figures they hold.
	card := cardNumbers([]string{
		"Value on 2025-10-31: 56,719.52 PLN gross | Net | 55,017.93 PLN",
		"Value on 2025-11-30: 58,686.37 PLN gross | Net | 56,925.78 PLN",
	})
	tests := []struct {
		name   string
		answer string
		stray  bool
	}{
		{name: "the difference of the two gross figures, as the real answers wrote it", answer: "an increase of 1,966.85 PLN"},
		{name: "the same difference rounded to the unit", answer: "an increase of about 1,967 PLN"},
		{name: "the difference in a Polish format", answer: "wzrost o 1 966,85 zł"},
		{name: "the difference with a sign", answer: "a change of -1,966.85 PLN"},
		{name: "a difference cut instead of rounded", answer: "an increase of about 1,966 PLN", stray: true},
		{name: "the difference with the wrong grosze", answer: "an increase of 1,966.00 PLN", stray: true},
		{name: "a round figure that is no rounding of it", answer: "an increase of 2,000 PLN", stray: true},
		{name: "off by one grosz", answer: "an increase of 1,966.86 PLN", stray: true},
		{name: "the sum of two card numbers", answer: "together 115,405.89 PLN"},
		{name: "a chain of three card numbers", answer: "that is 170,423.82 PLN", stray: true},
		{name: "a difference of a sum and a number", answer: "that is 59,611.36 PLN", stray: true},
		{name: "a percentage change of the two gross figures at 0.01", answer: "up 3.47%"},
		{name: "the same percentage rounded to a tenth", answer: "up 3.5%"},
		{name: "the percentage in a Polish format", answer: "wzrost o 3,47%"},
		{name: "a percentage off by 0.01", answer: "up 3.46%", stray: true},
		{name: "the percentage figure with no percent sign is no percentage change", answer: "up 3.47 PLN", stray: true},
		{name: "a percentage change of a percentage figure that is an amount", answer: "up 1,966.85%", stray: true},
		{name: "a number the cards don't give at all", answer: "about 777.77 PLN", stray: true},
		{name: "a scaled number is no derived one", answer: "about 2k PLN", stray: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stray := false
			for _, n := range numbersIn(tt.answer) {
				if !grounded(&n, card) && !derivedFrom(&n, card) {
					stray = true
				}
			}
			if stray != tt.stray {
				t.Errorf("%q: a number in no card and no one-step result = %v, want %v", tt.answer, stray, tt.stray)
			}
		})
	}
}

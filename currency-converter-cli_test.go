package main

import (
	"testing"
)

func TestCheckIsCurrency(t *testing.T) {
	arg := "USD"
	want := true
	cc := &CurrencyConvert{date: "latest"}
	res := cc.CheckIsCurrency(arg)
	if want != res {
		t.Fatalf(`CheckIsCurrency("USD") = %t, want match for %t, nil`, res, want)
	}
}

func TestConvertCurrency(t *testing.T) {
	from := "USD"
	to := "JPY"
	var amount float64 = 1
	want := 156.79775322
	cc := &CurrencyConvert{date: "2024-04-30"}
	res, err := cc.ConvertCurrency(amount, from, to)
	if res != want || err != nil {
		t.Fatalf(`ConvertCurrency(1, "USD", "JPY") = %f, %v, want match for %f, nil`, res, err, want)
	}
}

package service

import "github.com/shopspring/decimal"

func dec(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

package dto

import "github.com/shopspring/decimal"

type GoalRow struct {
	ID           int64
	Name         string
	TargetAmount decimal.Decimal `ts_type:"string"`
	TargetDate   string
}

type CreateGoalInput struct {
	Name         string
	TargetAmount decimal.Decimal `ts_type:"string"`
	TargetDate   string
}

type UpdateGoalInput struct {
	ID           int64
	Name         string
	TargetAmount decimal.Decimal `ts_type:"string"`
	TargetDate   string
}

type DeleteGoalInput struct {
	ID int64
}

type GoalMutationResult struct {
	ID int64
}

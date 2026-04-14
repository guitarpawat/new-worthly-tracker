package dto

type GoalRow struct {
	ID           int64
	Name         string
	TargetAmount float64
	TargetDate   string
}

type CreateGoalInput struct {
	Name         string
	TargetAmount float64
	TargetDate   string
}

type UpdateGoalInput struct {
	ID           int64
	Name         string
	TargetAmount float64
	TargetDate   string
}

type DeleteGoalInput struct {
	ID int64
}

type GoalMutationResult struct {
	ID int64
}

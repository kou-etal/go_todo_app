package task

type DueOption int

const (
	Due7Days   DueOption = 7
	Due30Days  DueOption = 30
	Due90Days  DueOption = 90
	Due365Days DueOption = 365
)

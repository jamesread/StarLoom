package store

import "time"

type ChoreRow struct {
	ID          int
	FamilyID    int
	StarChartID int
	Title       string
	StarReward  int
	WeekdayMask int
	Active      bool
	CreatedAt   string
}

type ChoreAssignmentRow struct {
	ID            int
	ChoreID       int
	ChildMemberID int
}

type ChorePauseRow struct {
	ID        int
	FamilyID  int
	StartDate string
	EndDate   string
	Reason    string
	CreatedAt string
}

type ChoreCompletionRow struct {
	ID             int
	AssignmentID   int
	CompletionDate string
	LedgerEntryID  *int
	CreatedAt      string
}

type ChoreWithAssignments struct {
	Chore       ChoreRow
	Assignments []ChoreAssignmentRow
}

type WeeklyChartCompletion struct {
	AssignmentID   int
	ChoreID        int
	ChildMemberID  int
	CompletionDate string
	StarsEarned    int
	LedgerEntryID  *int
}

func parseDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}

func WeekDates(weekStart string) ([]string, error) {
	return weekDates(weekStart)
}

func weekDates(weekStart string) ([]string, error) {
	start, err := parseDate(weekStart)
	if err != nil {
		return nil, err
	}
	dates := make([]string, 7)
	for i := 0; i < 7; i++ {
		dates[i] = start.AddDate(0, 0, i).Format("2006-01-02")
	}
	return dates, nil
}

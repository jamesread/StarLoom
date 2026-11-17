package store

import "time"

type ChoreRow struct {
	Title       string
	CreatedAt   string
	ID          int
	FamilyID    int
	StarChartID int
	StarReward  int
	WeekdayMask int
	SortOrder   int
	Active      bool
}

type ChoreAssignmentRow struct {
	ID            int
	ChoreID       int
	ChildMemberID int
}

type ChorePauseRow struct {
	StartDate string
	EndDate   string
	Reason    string
	CreatedAt string
	ID        int
	FamilyID  int
}

type ChoreCompletionRow struct {
	LedgerEntryID  *int
	CompletionDate string
	CreatedAt      string
	ID             int
	AssignmentID   int
}

type ChoreWithAssignments struct {
	Assignments []ChoreAssignmentRow
	Chore       ChoreRow
}

type WeeklyChartCompletion struct {
	LedgerEntryID  *int
	CompletionDate string
	AssignmentID   int
	ChoreID        int
	ChildMemberID  int
	StarsEarned    int
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
	for i := range 7 {
		dates[i] = start.AddDate(0, 0, i).Format("2006-01-02")
	}
	return dates, nil
}

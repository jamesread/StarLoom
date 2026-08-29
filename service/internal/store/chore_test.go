package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChoreWeekdayMask(t *testing.T) {
	require.Equal(t, 21, WeekdaysToMask([]int{1, 3, 5}))
	require.Equal(t, []int{1, 3, 5}, MaskToWeekdays(21))
	require.True(t, IsScheduledOnDate(21, "2026-08-24"))  // Monday
	require.False(t, IsScheduledOnDate(21, "2026-08-25")) // Tuesday
}

func TestChoreCompletionFlow(t *testing.T) {
	st := OpenMemory()
	ctx := context.Background()

	familyID, err := st.CreateFamily(ctx, "Test")
	require.NoError(t, err)
	parentID, err := st.CreateMember(ctx, familyID, "Parent", MemberRoleParent, nil, "")
	require.NoError(t, err)
	childID, err := st.CreateMember(ctx, familyID, "Child", MemberRoleChild, nil, "#3498db")
	require.NoError(t, err)

	choreID, err := st.CreateChore(ctx, familyID, "Make bed", 2, 127, []int{childID})
	require.NoError(t, err)

	assign, err := st.GetAssignment(ctx, choreID, childID)
	require.NoError(t, err)
	require.NotNil(t, assign)

	date := "2026-08-25"
	entryID, err := st.InsertLedgerEntry(ctx, StarLedgerRow{
		FamilyID: familyID, ChildMemberID: childID, Amount: 2,
		EntryType: LedgerTypeAward, Note: "Chore: Make bed",
		CreatedByMemberID: &parentID,
	})
	require.NoError(t, err)
	_, err = st.InsertChoreCompletion(ctx, assign.ID, date, entryID)
	require.NoError(t, err)

	balance, err := st.GetMemberBalance(ctx, childID)
	require.NoError(t, err)
	require.Equal(t, 2, balance)

	completions, err := st.ListCompletionsForWeek(ctx, familyID, "2026-08-25", "2026-08-31")
	require.NoError(t, err)
	require.Len(t, completions, 1)

	choreLedgerIDs, err := st.ListChoreLedgerEntryIDs(ctx, familyID)
	require.NoError(t, err)
	require.Len(t, choreLedgerIDs, 1)

	bonus, err := st.ListBonusStarsForWeek(ctx, familyID, "2026-08-25", "2026-08-31", choreLedgerIDs)
	require.NoError(t, err)
	require.Empty(t, bonus)

	_, err = st.InsertLedgerEntry(ctx, StarLedgerRow{
		FamilyID: familyID, ChildMemberID: childID, Amount: 3,
		EntryType: LedgerTypeAward, Note: "Great attitude",
		CreatedByMemberID: &parentID,
	})
	require.NoError(t, err)

	bonus, err = st.ListBonusStarsForWeek(ctx, familyID, "2026-08-25", "2026-08-31", choreLedgerIDs)
	require.NoError(t, err)
	require.NotEmpty(t, bonus)
}

func TestChorePause(t *testing.T) {
	st := OpenMemory()
	ctx := context.Background()

	familyID, err := st.CreateFamily(ctx, "Test")
	require.NoError(t, err)

	_, err = st.CreateChorePause(ctx, familyID, "2026-07-01", "2026-07-14", "Holiday")
	require.NoError(t, err)

	paused, err := st.IsDatePaused(ctx, familyID, "2026-07-10")
	require.NoError(t, err)
	require.True(t, paused)

	paused, err = st.IsDatePaused(ctx, familyID, "2026-08-01")
	require.NoError(t, err)
	require.False(t, paused)
}

package store

import (
	"context"
	"fmt"
	"sort"
)

type memoryChoreNotifications struct {
	nextID int
	rows   map[int]ChoreNotificationSubscriptionRow
}

func (m *Memory) notifyState() *memoryChoreNotifications {
	if m.choreNotifications == nil {
		m.choreNotifications = &memoryChoreNotifications{rows: map[int]ChoreNotificationSubscriptionRow{}}
	}
	return m.choreNotifications
}

func (m *Memory) ListChoreNotificationSubscriptions(_ context.Context, subscriberMemberID int) ([]ChoreNotificationSubscriptionRow, error) {
	st := m.notifyState()
	out := make([]ChoreNotificationSubscriptionRow, 0)
	for _, row := range st.rows {
		if row.SubscriberMemberID == subscriberMemberID {
			out = append(out, row)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		leftChild := childKey(out[i].ChildMemberID)
		rightChild := childKey(out[j].ChildMemberID)
		if leftChild != rightChild {
			return leftChild < rightChild
		}
		leftChore := choreKey(out[i].ChoreID)
		rightChore := choreKey(out[j].ChoreID)
		if leftChore != rightChore {
			return leftChore < rightChore
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (m *Memory) ReplaceChoreNotificationSubscriptions(_ context.Context, familyID, subscriberMemberID int, subs []ChoreNotificationSubscriptionRow) error {
	st := m.notifyState()
	for id, row := range st.rows {
		if row.SubscriberMemberID == subscriberMemberID {
			delete(st.rows, id)
		}
	}
	for i := range subs {
		st.nextID++
		st.rows[st.nextID] = ChoreNotificationSubscriptionRow{
			ID:                 st.nextID,
			FamilyID:           familyID,
			SubscriberMemberID: subscriberMemberID,
			ChildMemberID:      subs[i].ChildMemberID,
			ChoreID:            subs[i].ChoreID,
			CreatedAt:          familyNow(),
		}
	}
	return nil
}

func (m *Memory) MatchingChoreNotificationSubscribers(_ context.Context, familyID, childMemberID, choreID int) ([]int, error) {
	st := m.notifyState()
	seen := map[int]bool{}
	out := make([]int, 0)
	for _, row := range st.rows {
		if row.FamilyID != familyID {
			continue
		}
		if row.ChildMemberID != nil && *row.ChildMemberID != childMemberID {
			continue
		}
		if row.ChoreID != nil && *row.ChoreID != choreID {
			continue
		}
		if seen[row.SubscriberMemberID] {
			continue
		}
		seen[row.SubscriberMemberID] = true
		out = append(out, row.SubscriberMemberID)
	}
	sort.Ints(out)
	return out, nil
}

func childKey(id *int) string {
	if id == nil {
		return "any"
	}
	return fmt.Sprintf("%d", *id)
}

func choreKey(id *int) string {
	if id == nil {
		return "all"
	}
	return fmt.Sprintf("%d", *id)
}

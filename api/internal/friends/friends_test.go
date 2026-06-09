package friends

import (
	"testing"

	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/google/uuid"
)

func TestViewFor(t *testing.T) {
	viewer := uuid.New()
	profile := uuid.New()
	friendshipID := uuid.New()

	tests := []struct {
		name   string
		rec    *Record
		want   string
		wantID bool
	}{
		{
			name: "accepted",
			rec: &Record{
				ID: friendshipID, RequesterID: viewer, AddresseeID: profile, Status: "accepted",
			},
			want:   "friends",
			wantID: true,
		},
		{
			name: "pending outgoing",
			rec: &Record{
				ID: friendshipID, RequesterID: viewer, AddresseeID: profile, Status: "pending",
			},
			want:   "pending_outgoing",
			wantID: true,
		},
		{
			name: "pending incoming",
			rec: &Record{
				ID: friendshipID, RequesterID: profile, AddresseeID: viewer, Status: "pending",
			},
			want:   "pending_incoming",
			wantID: true,
		},
		{
			name: "own profile",
			rec: &Record{
				ID: friendshipID, RequesterID: viewer, AddresseeID: profile, Status: "accepted",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileID := profile
			viewerID := viewer
			if tt.name == "own profile" {
				profileID = viewer
			}

			got := ViewFor(viewerID, profileID, tt.rec)
			if tt.want == "" {
				if got != nil {
					t.Fatalf("expected nil view, got %#v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected friendship view")
			}
			if got.Status != tt.want {
				t.Fatalf("status = %q, want %q", got.Status, tt.want)
			}
			if tt.wantID && (got.ID == nil || *got.ID != friendshipID) {
				t.Fatalf("unexpected id %#v", got.ID)
			}
		})
	}
}

func TestViewForNilRecord(t *testing.T) {
	if got := ViewFor(uuid.New(), uuid.New(), nil); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestViewForDeclined(t *testing.T) {
	viewer := uuid.New()
	profile := uuid.New()
	got := ViewFor(viewer, profile, &Record{
		ID: uuid.New(), RequesterID: viewer, AddresseeID: profile, Status: "declined",
	})
	if got == nil || got.Status != "declined" {
		t.Fatalf("expected declined, got %#v", got)
	}
	_ = models.FriendshipView{}
}

package command

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCrewMembers(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		creator string
		want    []string
		wantErr string
	}{
		{
			name:    "spaces",
			input:   " bob ",
			creator: "alice",
			want:    []string{"alice", "bob"},
		},
		{
			name:    "commas",
			input:   "bob, alice",
			creator: "alice",
			want:    []string{"alice", "bob"},
		},
		{
			name:    "mixed separators",
			input:   "alice, bob",
			creator: "alice",
			want:    []string{"alice", "bob"},
		},
		{
			name:    "creator listed explicitly",
			input:   "alice, bob",
			creator: "alice",
			want:    []string{"alice", "bob"},
		},
		{
			name:    "creator listed with spaces and commas",
			input:   "alice, alice, bob",
			creator: "alice",
			want:    []string{"alice", "bob"},
		},
		{
			name:    "creator only",
			input:   "alice",
			creator: "alice",
			wantErr: "at least 2 unique",
		},
		{
			name:    "duplicate additional member",
			input:   "bob bob",
			creator: "alice",
			wantErr: "Duplicate crew member",
		},
		{
			name:    "invalid member",
			input:   "not-valid",
			creator: "alice",
			wantErr: "Invalid Telegram username",
		},
		{
			name:    "empty comma-separated member",
			input:   "bob,,",
			creator: "alice",
			wantErr: "must be separated",
		},
		{
			name:    "more than one additional member",
			input:   "bob,carol",
			creator: "alice",
			wantErr: "only one additional",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCrewMembers(test.input, test.creator)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("parseCrewMembers() error = %v", err)
				}
				if !reflect.DeepEqual(got, test.want) {
					t.Fatalf("parseCrewMembers() = %v, want %v", got, test.want)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("parseCrewMembers() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

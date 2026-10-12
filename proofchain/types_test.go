package proofchain

import (
	"encoding/json"
	"testing"
)

func TestTimestampUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string // raw JSON value including quotes where relevant
		wantYear  int
		wantMonth int
		wantDay   int
		wantZero  bool
		wantErr   bool
	}{
		{name: "RFC3339 with zone", input: `"2026-07-21T10:30:00Z"`, wantYear: 2026, wantMonth: 7, wantDay: 21},
		{name: "RFC3339Nano", input: `"2026-07-21T10:30:00.123456789Z"`, wantYear: 2026, wantMonth: 7, wantDay: 21},
		{name: "no-zone microseconds", input: `"2026-07-21T10:30:00.123456"`, wantYear: 2026, wantMonth: 7, wantDay: 21},
		{name: "no-zone seconds", input: `"2026-07-21T10:30:00"`, wantYear: 2026, wantMonth: 7, wantDay: 21},
		{name: "date only", input: `"2026-07-21"`, wantYear: 2026, wantMonth: 7, wantDay: 21},
		{name: "empty string is zero value", input: `""`, wantZero: true},
		{name: "literal null is zero value", input: `"null"`, wantZero: true},
		{name: "unparseable is error", input: `"not-a-date"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts Timestamp
			err := json.Unmarshal([]byte(tt.input), &ts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if tt.wantZero {
				if !ts.Time.IsZero() {
					t.Errorf("expected zero time for %q, got %v", tt.input, ts.Time)
				}
				return
			}
			if ts.Year() != tt.wantYear || int(ts.Month()) != tt.wantMonth || ts.Day() != tt.wantDay {
				t.Errorf("parsed %q = %04d-%02d-%02d, want %04d-%02d-%02d",
					tt.input, ts.Year(), int(ts.Month()), ts.Day(), tt.wantYear, tt.wantMonth, tt.wantDay)
			}
		})
	}
}

func TestTimestampUnmarshalWithinStruct(t *testing.T) {
	// Verifies the custom Timestamp works when embedded in a typed payload.
	payload := `{"channel_id":"ch1","name":"n","state":"open","created_at":"2026-01-02T03:04:05Z"}`
	var ch Channel
	if err := json.Unmarshal([]byte(payload), &ch); err != nil {
		t.Fatalf("unmarshal Channel failed: %v", err)
	}
	if ch.State != ChannelStateOpen {
		t.Errorf("state = %q, want %q", ch.State, ChannelStateOpen)
	}
	if ch.CreatedAt.Year() != 2026 || ch.CreatedAt.Day() != 2 {
		t.Errorf("created_at parsed incorrectly: %v", ch.CreatedAt.Time)
	}
}

func TestEnumConstants(t *testing.T) {
	// Guards against accidental string-value drift on the wire-facing enums.
	tests := []struct {
		got  string
		want string
	}{
		{got: string(AttestationModeDirect), want: "direct"},
		{got: string(AttestationModeBatch), want: "batch"},
		{got: string(AttestationModeChannel), want: "channel"},
		{got: string(EventStatusPending), want: "pending"},
		{got: string(EventStatusConfirmed), want: "confirmed"},
		{got: string(EventStatusSettled), want: "settled"},
		{got: string(EventStatusFailed), want: "failed"},
		{got: string(ChannelStateOpen), want: "open"},
		{got: string(ChannelStateSettled), want: "settled"},
		{got: string(ChannelStateClosed), want: "closed"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("enum value = %q, want %q", tt.got, tt.want)
		}
	}
}

package command

import "testing"

func TestDecide(t *testing.T) {
	tests := []struct {
		name string
		c    Classification
		want Decision
	}{
		{"read is auto", Classification{Mode: ModeRead}, DecisionAutoApprove},
		{"mutation needs approval", Classification{Mode: ModeMutation}, DecisionRequireApproval},
		{"unknown/zero mode is read by zero value (drivers must set mutation explicitly)", Classification{}, DecisionAutoApprove},
		{"deny wins over read", Classification{Mode: ModeRead, DenyReason: "blocked"}, DecisionDeny},
		{"deny wins over mutation", Classification{Mode: ModeMutation, DenyReason: "blocked"}, DecisionDeny},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Decide(tt.c); got != tt.want {
				t.Fatalf("Decide(%+v) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}

func TestClassificationIsDenied(t *testing.T) {
	if (Classification{Mode: ModeMutation}).IsDenied() {
		t.Fatal("mutation without reason must not be denied")
	}
	if !(Classification{Mode: ModeMutation, DenyReason: "x"}).IsDenied() {
		t.Fatal("reason must mark denied")
	}
}

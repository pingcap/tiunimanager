package structs

import (
	"testing"
	"time"
)

func TestToken_IsValid(t *testing.T) {
	type fields struct {
		TokenString    string
		AccountName    string
		AccountID      string
		TenantID       string
		TenantName     string
		ExpirationTime time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{
				TokenString:    tt.fields.TokenString,
				AccountName:    tt.fields.AccountName,
				AccountID:      tt.fields.AccountID,
				TenantID:       tt.fields.TenantID,
				TenantName:     tt.fields.TenantName,
				ExpirationTime: tt.fields.ExpirationTime,
			}
			if got := token.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

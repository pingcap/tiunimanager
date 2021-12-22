package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/models/common"
	"reflect"
	"testing"
	"time"
)

func TestTokenReadWrite_AddToken(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx            context.Context
		tokenString    string
		accountName    string
		accountId      string
		tenantId       string
		expirationTime time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &TokenReadWrite{
				GormDB: tt.fields.GormDB,
			}
			got, err := g.AddToken(tt.args.ctx, tt.args.tokenString, tt.args.accountName, tt.args.accountId, tt.args.tenantId, tt.args.expirationTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

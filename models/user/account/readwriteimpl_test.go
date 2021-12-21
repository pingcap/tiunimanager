package account

import (
	"context"
	"github.com/pingcap-inc/tiem/models/common"
	"reflect"
	"testing"
)

func TestAccountReadWrite_AddAccount(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx       context.Context
		tenantId  string
		name      string
		salt      string
		finalHash string
		status    int8
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &AccountReadWrite{
				GormDB: tt.fields.GormDB,
			}
			got, err := g.AddAccount(tt.args.ctx, tt.args.tenantId, tt.args.name, tt.args.salt, tt.args.finalHash, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

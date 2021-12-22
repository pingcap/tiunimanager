package tenant

import (
	"context"
	"github.com/pingcap-inc/tiem/message"
	"reflect"
	"testing"
)

func TestManager_CreateTenant(t *testing.T) {
	type args struct {
		ctx     context.Context
		request message.CreateTenantReq
	}
	tests := []struct {
		name    string
		args    args
		want    *message.CreateTenantResp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Manager{}
			got, err := p.CreateTenant(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}

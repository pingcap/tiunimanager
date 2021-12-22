package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/message"
	"reflect"
	"testing"
)

func TestManager_Login(t *testing.T) {
	type args struct {
		ctx     context.Context
		request message.LoginReq
	}
	tests := []struct {
		name     string
		args     args
		wantResp message.LoginResp
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Manager{}
			gotResp, err := p.Login(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("Login() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

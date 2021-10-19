package framework

import (
	 "github.com/pingcap-inc/tiem/library/common"
	"reflect"
	"testing"
)

func TestNewCertificateFromArgs(t *testing.T) {
	type args struct {
		args *ClientArgs
	}
	tests := []struct {
		name string
		args args
		want *CertificateInfo
	}{
		{"normal", args{&ClientArgs{DeployDir: "aaaa"}}, &CertificateInfo{
			CertificateCrtFilePath: "aaaa" + common.CertDirPrefix + common.CrtFileName,
			CertificateKeyFilePath: "aaaa" + common.CertDirPrefix + common.KeyFileName,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCertificateFromArgs(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCertificateFromArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

package knowledge

import (
	"reflect"
	"testing"
)

func TestClusterComponentFromCode(t *testing.T) {
	t.Run("TiDB", func(t *testing.T) {
		got := ClusterComponentFromCode("TiDB")
		if got.ComponentType != "TiDB" {
			t.Errorf("ClusterComponentFromCode() = %v, want code = %v", got, "TiDB")
		}
	})
	t.Run("TiKV", func(t *testing.T) {
		got := ClusterComponentFromCode("TiKV")
		if got.ComponentType != "TiKV" {
			t.Errorf("ClusterComponentFromCode() = %v, want code = %v", got, "TiKV")
		}
	})
	t.Run("PD", func(t *testing.T) {
		got := ClusterComponentFromCode("PD")
		if got.ComponentType != "PD" {
			t.Errorf("ClusterComponentFromCode() = %v, want code = %v", got, "PD")
		}
	})
	t.Run("nil", func(t *testing.T) {
		got := ClusterComponentFromCode("sss")
		if got != nil {
			t.Errorf("ClusterComponentFromCode() = %v, want nil", got)
		}
	})
}

func TestClusterTypeFromCode(t *testing.T) {
	t.Run("TiDB", func(t *testing.T) {
		got := ClusterTypeFromCode("TiDB")
		if got.Code != "TiDB" {
			t.Errorf("ClusterTypeFromCode() = %v, want code = %v", got, "TiDB")
		}
	})
	t.Run("nil", func(t *testing.T) {
		got := ClusterTypeFromCode("wwww")
		if got != nil {
			t.Errorf("ClusterTypeFromCode() = %v, want nil", got)
		}
	})
}

func TestClusterVersionFromCode(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name string
		args args
		want *ClusterVersion
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClusterVersionFromCode(tt.args.code); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClusterVersionFromCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadKnowledge(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestParameterFromName(t *testing.T) {
	got := ParameterFromName("binlog_cache_size")
	if got.Name != "binlog_cache_size" {
		t.Errorf("ParameterFromName() = %v, want %v", got, "binlog_cache_size")
	}
}

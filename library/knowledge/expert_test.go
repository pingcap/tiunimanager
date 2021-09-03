package knowledge

import (
	"reflect"
	"testing"
)

func TestClusterComponentFromCode(t *testing.T) {
	t.Run("tidb", func(t *testing.T) {
		got := ClusterComponentFromCode("tidb")
		if got.ComponentType != "tidb" {
			t.Errorf("ClusterComponentFromCode() = %v, want code = %v", got, "tidb")
		}
	})
	t.Run("tidb", func(t *testing.T) {
		got := ClusterComponentFromCode("tikv")
		if got.ComponentType != "tikv" {
			t.Errorf("ClusterComponentFromCode() = %v, want code = %v", got, "tikv")
		}
	})
	t.Run("tidb", func(t *testing.T) {
		got := ClusterComponentFromCode("pd")
		if got.ComponentType != "pd" {
			t.Errorf("ClusterComponentFromCode() = %v, want code = %v", got, "pd")
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
	t.Run("tidb", func(t *testing.T) {
		got := ClusterTypeFromCode("tidb")
		if got.Code != "tidb" {
			t.Errorf("ClusterTypeFromCode() = %v, want code = %v", got, "tidb")
		}
	})
	t.Run("dm", func(t *testing.T) {
		got := ClusterTypeFromCode("dm")
		if got.Code != "dm" {
			t.Errorf("ClusterTypeFromCode() = %v, want code = %v", got, "dm")
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

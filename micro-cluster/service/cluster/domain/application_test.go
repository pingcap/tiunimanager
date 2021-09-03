package domain

import "testing"

func Test_buildConfig(t *testing.T) {
	type args struct {
		task    *TaskEntity
		context *FlowContext
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildConfig(tt.args.task, tt.args.context); got != tt.want {
				t.Errorf("buildConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

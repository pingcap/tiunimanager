package domain

import "testing"

func TestClusterStatusFromValue(t *testing.T) {
	type args struct {
		v int
	}
	tests := []struct {
		name string
		args args
		want ClusterStatus
	}{
		{"ClusterStatusUnlined", args{0}, ClusterStatusUnlined},
		{"ClusterStatusOnline", args{1}, ClusterStatusOnline},
		{"ClusterStatusOffline", args{2}, ClusterStatusOffline},
		{"ClusterStatusDeleted", args{3}, ClusterStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClusterStatusFromValue(tt.args.v); got != tt.want {
				t.Errorf("ClusterStatusFromValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClusterStatus_Display(t *testing.T) {
	tests := []struct {
		name string
		s    ClusterStatus
		want string
	}{
		{"ClusterStatusUnlined", ClusterStatusUnlined, "未上线"},
		{"ClusterStatusOnline", ClusterStatusOnline, "运行中"},
		{"ClusterStatusOffline", ClusterStatusOffline, "已下线"},
		{"ClusterStatusDeleted", ClusterStatusDeleted, "已删除"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Display(); got != tt.want {
				t.Errorf("Display() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatusFromValue(t *testing.T) {
	type args struct {
		v int
	}
	tests := []struct {
		name string
		args args
		want TaskStatus
	}{
		{"TaskStatusInit", args{0}, TaskStatusInit},
		{"TaskStatusProcessing", args{1}, TaskStatusProcessing},
		{"TaskStatusFinished", args{2}, TaskStatusFinished},
		{"TaskStatusError", args{3}, TaskStatusError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TaskStatusFromValue(tt.args.v); got != tt.want {
				t.Errorf("TaskStatusFromValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_Display(t *testing.T) {
	tests := []struct {
		name string
		s    TaskStatus
		want string
	}{
		{"TaskStatusInit", TaskStatusInit, "未开始"},
		{"TaskStatusProcessing", TaskStatusProcessing, "进行中"},
		{"TaskStatusFinished", TaskStatusFinished, "完成"},
		{"TaskStatusError", TaskStatusError, "失败"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Display(); got != tt.want {
				t.Errorf("Display() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_Finished(t *testing.T) {
	tests := []struct {
		name string
		s    TaskStatus
		want bool
	}{
		{"TaskStatusInit", TaskStatusInit, false},
		{"TaskStatusProcessing", TaskStatusProcessing, false},
		{"TaskStatusFinished", TaskStatusFinished, true},
		{"TaskStatusError", TaskStatusError, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Finished(); got != tt.want {
				t.Errorf("Finished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkBackupRangeValid(t *testing.T) {
	type args struct {
		backupRange string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"BackupRangeIncrement", args{"incr"}, true},
		{"BackupRangeFull", args{"full"}, true},
		{"whatever", args{"whatever"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkBackupRangeValid(tt.args.backupRange); got != tt.want {
				t.Errorf("checkBackupRangeValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkBackupTypeValid(t *testing.T) {
	type args struct {
		backupType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"BackupTypeLogic", args{"logical"}, true},
		{"BackupTypePhysics", args{"physical"}, true},
		{"whatever", args{"whatever"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkBackupTypeValid(tt.args.backupType); got != tt.want {
				t.Errorf("checkBackupTypeValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkWeekDayValid(t *testing.T) {
	type args struct {
		weekday string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Sunday", args{"Sunday"}, true},
		{"Monday", args{"Monday"}, true},
		{"Tuesday", args{"Tuesday"}, true},
		{"Wednesday", args{"Wednesday"}, true},
		{"Thursday", args{"Thursday"}, true},
		{"Friday", args{"Friday"}, true},
		{"Saturday", args{"Saturday"}, true},
		{"whatever", args{"whatever"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkWeekDayValid(tt.args.weekday); got != tt.want {
				t.Errorf("checkWeekDayValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

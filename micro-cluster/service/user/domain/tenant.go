package domain

// Tenant 租户
type Tenant struct {
	Name   string
	Id     string
	Type   TenantType
	Status CommonStatus
}

type TenantType int

const (
	SystemManagement  TenantType = 0
	InstanceWorkspace TenantType = 1
	PluginAccess      TenantType = 2
)

type CommonStatus int

const (
	Valid              CommonStatus = 0
	Invalid            CommonStatus = 1
	Deleted            CommonStatus = 2
	UnrecognizedStatus CommonStatus = -1
)

func (s CommonStatus) IsValid() bool {
	return s == Valid
}

func CommonStatusFromStatus(status int32) CommonStatus {
	switch status {
	case 0:
		return Valid
	case 1:
		return Invalid
	case 2:
		return Deleted
	default:
		return UnrecognizedStatus
	}
}

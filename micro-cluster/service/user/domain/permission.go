package domain

type Permission struct {
	TenantId string
	Code     string
	Name     string
	Type     PermissionType
	Desc     string
	Status   CommonStatus
}

type PermissionType int

type PermissionAggregation struct {
	Permission
	Roles []Role
}

const (
	UnrecognizedType PermissionType = 0
	Path             PermissionType = 1
	Act              PermissionType = 2
	Data             PermissionType = 3
)

func PermissionTypeFromType(pType int32) PermissionType {
	switch pType {
		case 1: return Path
		case 2: return Act
		case 3: return Data
		default: return UnrecognizedType
	}
}

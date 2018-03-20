package utils

// ParamArg struct that keep the param flags
type ParamArg struct {
	StorageClass   string
	PermissionName string

	IsVerbose bool
	IsPublic  bool
}

// GlobalParamArg global struct that keep the param flags
var GlobalParamArg ParamArg

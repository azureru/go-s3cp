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

// GlobalPermissions is generic permissions that applicable in many providers
var GlobalPermissions = []string{"private", "public-read", "public-read-write", "authenticated-read", "bucket-owner-read", "bucket-owner-full-control"}

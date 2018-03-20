package utils

import (
	"errors"
	"strings"
)

// IsGCS will return true if the path is a GCS path gs://bucket/path
func IsGCS(source string) bool {
	if strings.HasPrefix(source, "gs://") {
		return true
	}
	return false
}

// ShortGCSExtract will extract GCS short path into bucket, path string
func ShortGCSExtract(source string) (string, string, error) {
	if !IsGCS(source) {
		return "", "", errors.New("Invalid GCS short-path")
	}

	// gcs path is gs://bucket/path/path/path
	source = strings.Replace(source, "gs://", "", 1)
	args := strings.Split(source, "/")
	if len(args) < 2 {
		return "", "", errors.New("Invalid GCS short-path")
	}
	bucket := args[0]
	path := strings.Join(args[1:], "")

	return bucket, path, nil
}

// UploadGCS to be used to upload local-file to target s3 path in
// our short notation gs://bucket/path
func UploadGCS(localSource string, remoteTarget string) error {

	return nil
}

// DownloadGCS to download GCS file from remote source to local file target
func DownloadGCS(remoteSource string, localTarget string) error {

	return nil
}

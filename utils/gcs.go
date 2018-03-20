package utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
)

var gcsPredefinedStorageClasses = []string{
	"coldline",
	"hardline",
	"regional",
	"multi_regional",
}

var gcsPredefinedPermissionNames = []string{
	"authenticated-write",
	"authenticated-read",
	"public-read",
	"public-read-write",
}

// GCSCredentialSessionDefined when true that means the credential for GCS is defined
var GCSCredentialSessionDefined bool

// GCSCredentialSession the GCS credential session
var GCSCredentialSession context.Context

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
func UploadGCS(localSource string, remoteTarget string, permissionName string, storageClass string) error {
	log.Println("Uploading ", localSource, " to ", remoteTarget)

	isSourceIsFolder := IsDir(localSource)

	isTargetIsFolder := false
	if strings.HasSuffix(remoteTarget, "/") {
		isTargetIsFolder = true
	}

	bucket, destPath, err := ShortGCSExtract(remoteTarget)
	if err != nil {
		return errors.New("Invalid Bucket Path")
	}

	// if credential not defined beforehand - we use default credential (the one from ENV var)
	if GCSCredentialSessionDefined == false {
		GCSCredentialSessionDefined = true
		GCSCredentialSession = context.Background()
	}

	client, err := storage.NewClient(GCSCredentialSession)
	if err != nil {
		return err
	}

	bucketGcs := client.Bucket(bucket)
	acl := normalizePermission(permissionName)
	storageClass = normalizeStorageClass(storageClass)

	// walk the files
	walker := make(FileWalk)
	go func() {
		// Gather the files to upload by walking the path recursively.
		if err := filepath.Walk(localSource, walker.Walk); err != nil {
			log.Fatalln("Error: ", fmt.Sprintf("%v", err))
		}
		close(walker)
	}()
	// For each file found on the recursive
	for path := range walker {
		rel, err := filepath.Rel(localSource, path)
		if err != nil {
			log.Fatalln("Unable to get relative path:", path, err)
		}
		file, err := os.Open(path)
		if err != nil {
			log.Println("Failed opening file", path, err)
			continue
		}
		defer file.Close()

		targetKey := remoteTarget
		prefix := ""
		if isTargetIsFolder {
			prefix = destPath
		}
		if isSourceIsFolder {
			targetKey = filepath.Join(prefix, rel)
		}
		fmt.Println(" ..." + path + " to " + targetKey)

		wc := bucketGcs.Object(targetKey).NewWriter(GCSCredentialSession)

		wc.CacheControl = "max-age=86400"
		wc.StorageClass = storageClass
		wc.ACL = acl

		// writing
		slurp, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		buffer := slurp[0:512]
		wc.ContentType = http.DetectContentType(buffer)

		if _, err := wc.Write(slurp); err != nil {
			return err
		}
		if err := wc.Close(); err != nil {
			return err
		}
		slurp = nil
	}

	return nil
}

// DownloadGCS to download GCS file from remote source to local file target
func DownloadGCS(remoteSource string, localTarget string) error {

	log.Println("Downloading ", remoteSource, " to ", localTarget)
	bucket, path, err := ShortGCSExtract(remoteSource)
	if err != nil {
		return errors.New("Invalid Bucket Path")
	}

	// if credential not defined beforehand - we use default credential (the one from ENV var)
	if GCSCredentialSessionDefined == false {
		GCSCredentialSessionDefined = true
		GCSCredentialSession = context.Background()
	}

	client, err := storage.NewClient(GCSCredentialSession)
	if err != nil {
		return err
	}

	// download start!
	rc, err := client.Bucket(bucket).Object(path).NewReader(GCSCredentialSession)
	if err != nil {
		return err
	}
	defer rc.Close()

	// download successfull - store to file
	err = WriteReaderToFile(rc, localTarget)
	if err != nil {
		return err
	}

	return nil
}

// normalizeStorageClass will normalize sc string to proper storage class
func normalizeStorageClass(sc string) string {
	if sc == "" {
		sc = "multi_regional"
	}
	return sc
}

// normalizePermission will normalize pm string to proper ACL rule
func normalizePermission(pm string) []storage.ACLRule {
	if pm == "" {
		pm = "authenticated-read"
	}

	if pm == "public-read" {
		// public-read All user can read
		return []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	} else if pm == "public-read-write" {
		// public-read-write all-user can write
		return []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleWriter}}
	} else if pm == "authenticated-read" {
		// auth read
		return []storage.ACLRule{{Entity: storage.AllAuthenticatedUsers, Role: storage.RoleReader}}
	} else if pm == "authenticated-write" {
		// auth write
		return []storage.ACLRule{{Entity: storage.AllAuthenticatedUsers, Role: storage.RoleWriter}}
	}

	// default
	return []storage.ACLRule{{Entity: storage.AllAuthenticatedUsers, Role: storage.RoleReader}}
}

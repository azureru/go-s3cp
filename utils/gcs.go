package utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
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
	path := strings.Join(args[1:], "/")

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
	storageClass = normalizeStorageClass(storageClass)

	// walk the files
	walker := make(FileWalk)
	walker.IterateUpload(localSource, func(path, rel string) error {
		file, err := os.Open(path)
		if err != nil {
			log.Println("Failed opening file", path, err)
			return nil
		}
		defer file.Close()

		var targetKey string
		prefix := ""
		if isTargetIsFolder {
			prefix = destPath
		}
		if isSourceIsFolder {
			targetKey = filepath.Join(prefix, rel)
		} else {
			if !isTargetIsFolder {
				targetKey = destPath
			} else {
				if rel == "." {
					rel = path
				}
				targetKey = filepath.Join(prefix, rel)
			}
		}
		fmt.Println(" ..." + path + " to " + targetKey)

		obj := bucketGcs.Object(targetKey)
		wc := obj.NewWriter(GCSCredentialSession)
		wc.CacheControl = "max-age=86400"
		wc.StorageClass = storageClass
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

		// set the object permission
		if err := setPermission(GCSCredentialSession, obj.ACL(), permissionName); err != nil {
			return err
		}

		return nil
	})

	return nil
}

// DownloadGCS to download GCS file from remote source to local file target
func DownloadGCS(remoteSource string, localTarget string) error {

	isSourceIsFolder := strings.HasSuffix(remoteSource, "/")
	if isSourceIsFolder {
		log.Fatalln("Remote source must be a file")
	}

	log.Println("Downloading ", remoteSource, " to ", localTarget)
	bucket, sourcePath, err := ShortGCSExtract(remoteSource)
	if err != nil {
		log.Fatalln("Invalid Bucket Path ", remoteSource)
	}
	fmt.Println(" " + bucket + " " + sourcePath)

	filename := localTarget
	isTargetIsFolder := IsDir(localTarget)
	if isTargetIsFolder {
		file := path.Base(sourcePath)
		filename = localTarget + file
	}

	// if credential not defined beforehand - we use default credential (the one from ENV var)
	if GCSCredentialSessionDefined == false {
		GCSCredentialSessionDefined = true
		GCSCredentialSession = context.Background()
	}

	client, err := storage.NewClient(GCSCredentialSession)
	if err != nil {
		log.Fatalln(err)
	}

	// download start!
	rc, err := client.Bucket(bucket).Object(sourcePath).NewReader(GCSCredentialSession)
	if err != nil {
		log.Fatalln(err)
	}
	defer rc.Close()

	// download successfull - store to file
	err = WriteReaderToFile(rc, filename)
	if err != nil {
		log.Fatalln(err)
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

	// default is authenticated-read
	return []storage.ACLRule{{Entity: storage.AllAuthenticatedUsers, Role: storage.RoleReader}}
}

// normalizePermission will normalize pm string to proper ACL rule
func setPermission(ctx context.Context, acl *storage.ACLHandle, pm string) error {

	if pm == "public-read" {
		// public-read All user can read
		return acl.Set(ctx, storage.AllUsers, storage.RoleReader)
	} else if pm == "public-read-write" {
		// public-read-write all-user can write
		return acl.Set(ctx, storage.AllUsers, storage.RoleWriter)
	} else if pm == "authenticated-read" {
		// auth read
		return acl.Set(ctx, storage.AllAuthenticatedUsers, storage.RoleReader)
	} else if pm == "authenticated-write" {
		// auth write
		return acl.Set(ctx, storage.AllAuthenticatedUsers, storage.RoleWriter)
	}

	// default is no permission 
	return nil
}

package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var s3predefinedPermissions = []string{"private", "public-read", "public-read-write", "authenticated-read", "bucket-owner-read", "bucket-owner-full-control"}
var s3predefinedRegions = []string{"us-east-1", "us-west-1", "us-west-2", "eu-west-1", "eu-central-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "sa-east-1"}

// S3Region global default region
var S3Region string

// UploadS3 to be used to upload local-file to target s3 path in
// our short notation s3:region:bucket:path or s3:bucket:path
func UploadS3(localSource string, remoteTarget string, permissionName string, storageClass string) error {
	isSourceIsFolder := IsDir(localSource)

	isTargetIsFolder := false
	if strings.HasSuffix(remoteTarget, "/") {
		isTargetIsFolder = true
	}

	if permissionName == "" {
		permissionName = "private"
	}
	if storageClass == "" {
		storageClass = "STANDARD"
	}

	// extract the target s3 short path
	region, bucket, destPath, err := ShortS3Extract(remoteTarget)
	if err != nil {
		log.Fatalln("Error: ", fmt.Sprintf("%v", err))
	}
	S3Region = region

	// check for region
	if !stringInSlice(region, s3predefinedRegions) {
		log.Fatalln("Error: Invalid Region Name")
	}

	// prepare a session manager
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(S3Region)}))
	uploader := s3manager.NewUploader(sess)

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

		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:       aws.String(bucket),
			Key:          aws.String(targetKey),
			ACL:          aws.String(permissionName),
			Body:         file,
			StorageClass: aws.String(storageClass),
		})
		if err != nil {
			log.Fatalln("Failed to upload", path, err)
		}
		fmt.Println("    Uploaded: ", result.Location)
	}

	return nil
}

// IsS3 extract the path and identify whether it's s3 short path or not
func IsS3(source string) bool {
	_, _, _, err := ShortS3Extract(source)
	if err != nil {
		return false
	}

	return true
}

// ShortS3Extract to extract our short s3 path to region, bucket, path components
func ShortS3Extract(source string) (string, string, string, error) {

	args := strings.Split(source, ":")
	// s3:region:bucket:path
	if len(args) < 4 {
		return "", "", "", errors.New("S3 path must be in format s3:region:bucketname:targetpath")
	}
	s3Prefix := args[0]
	region := args[1]
	bucket := args[2]
	destPath := args[3]

	if s3Prefix != "s3" {
		return "", "", "", errors.New("S3 path must be in format s3:region:bucketname:targetpath")
	}

	return region, bucket, destPath, nil
}

// DownloadS3 to download s3 file from remote source to local file target
func DownloadS3(remoteSource string, localTarget string) error {

	isSourceIsFolder := strings.HasSuffix(remoteSource, "/")
	region, bucket, sourcePath, err := ShortS3Extract(remoteSource)
	if err != nil {
		log.Fatalln("Error: ", fmt.Sprintf("%v", err))
	}
	if isSourceIsFolder {
		log.Fatalln("Remote source must be a file")
	}

	// check for region
	if !stringInSlice(region, s3predefinedRegions) {
		log.Fatalln("Invalid Region Name " + region)
	}
	S3Region = region

	filename := localTarget
	isTargetIsFolder := IsDir(localTarget)
	if isTargetIsFolder {
		file := path.Base(sourcePath)
		filename = localTarget + file
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(S3Region)}))
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create file %q, %v", filename, err)
	}
	defer f.Close()

	// Write the contents of S3 Object to the file
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(sourcePath),
	})
	if err != nil {
		// fail - dispose garbage
		os.Remove(filename)
		log.Fatalf("failed to download file, %v", err)
	}
	fmt.Printf("File downloaded, %d bytes\n", n)

	return nil
}

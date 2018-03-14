package main

import (
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
	"github.com/codegangsta/cli"
)

var (
	firstPath  string
	secondPath string
	prefix     string
	sourcePath string

	region   string
	bucket   string
	destPath string

	// storage class to be used (default STANDARD)
	storageClass string

	// the permission of the uploaded file (default authenticated-read)
	permissionName string

	// true when source is a folder (so we need to walk on it, false when not)
	isSourceIsFolder bool

	// true when target is a folder (end with / e.g. south:bucket:/root/one/wise/, false when not - will be assumed as filename)
	isTargetIsFolder bool

	// by default all uploaded file will be `private` you need to specify public flag
	isPublic bool

	// verbose mode
	isVerbose bool
)

var predefinedPermissions = []string{"private", "public-read", "public-read-write", "authenticated-read", "bucket-owner-read", "bucket-owner-full-control"}
var predefinedRegions = []string{"us-east-1", "us-west-1", "us-west-2", "eu-west-1", "eu-central-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "sa-east-1"}

func main() {
	app := cli.NewApp()
	app.Name = "go-s3cp"
	app.Author = "Erwin Saputra"
	app.Version = "0.0.1"
	app.Usage = "Copy files from local to s3"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, vv",
			Usage: "show Verbose Output",
		},
		cli.BoolFlag{
			Name:  "public, p",
			Usage: "upload as Public ACL, this will ignore setting on permission",
		},
		cli.StringFlag{
			Name:  "storage, s",
			Usage: "specify the storage class (STANDARD | REDUCED_REDUNDANCY | STANDARD_IA)",
		},
		cli.StringFlag{
			Name:  "permission, perm",
			Usage: "specify the Permission of the file (private | public-read | public-read-write | authenticated-read | bucket-owner-read | bucket-owner-full-control)",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "region",
			Aliases: []string{"reg"},
			Usage:   "List all Amazon Regions (as a reminder :P)",
			Action: func(c *cli.Context) {
				// since the SDK does not provide ways to enumerate this -
				fmt.Println("List of Regions:")
				for index := 0; index < len(predefinedRegions); index++ {
					fmt.Printf("   %s\n", predefinedRegions[index])
				}
			},
		},
		{
			Name:  "permission",
			Usage: "List all Permissions",
			Action: func(c *cli.Context) {
				fmt.Println("List of Permissions:")
				for index := 0; index < len(predefinedPermissions); index++ {
					fmt.Printf("   %s\n", predefinedPermissions[index])
				}
			},
		},
	}

	app.Action = func(c *cli.Context) {
		args := c.Args()
		if len(args) < 2 {
			fmt.Println("You need to put [from] and [to] path.")
			fmt.Println("EXAMPLE:")
			fmt.Println("   go-s3cp ./file regionname:bucket:path/path/filename")
			fmt.Println("   go-s3cp ./file/ \"regionname:bucket:path/path with spaces/filename/\"")
			return
		}
		firstPath = args[0]
		secondPath = args[1]

		// storage class
		storageClass = "STANDARD"
		if c.String("storage") != "" {
			storageClass = c.String("storage")
		}
		fmt.Println("Storage Class: " + storageClass)

		isVerbose = c.Bool("verbose")

		// permission
		permissionName = "private"
		if c.String("permission") != "" {
			permissionName = c.String("permission")
		}
		isPublic = c.Bool("public")
		if isPublic {
			permissionName = "public-read"
		}
		if !stringInSlice(permissionName, predefinedPermissions) {
			log.Fatalln("Invalid Permission Name " + permissionName)
		}
		fmt.Println("Permission: " + permissionName)

		if strings.Contains(firstPath, ":") {
			copyRemoteToLocal()
		} else {
			copyLocalToRemote()
		}
	}

	app.Run(os.Args)
}

func copyRemoteToLocal() {
	// parse target path
	args := strings.Split(firstPath, ":")
	isSourceIsFolder = strings.HasSuffix(firstPath, "/")
	// region:bucket:path
	if len(args) < 3 {
		log.Fatalln("S3 path must be in format region:bucketname:targetpath")
	}
	region = args[0]
	bucket = args[1]
	sourcePath = args[2]

	if isSourceIsFolder {
		log.Fatalln("Remote source must be a file")
	}

	// check for region
	if !stringInSlice(region, predefinedRegions) {
		log.Fatalln("Invalid Region Name " + region)
	}

	filename := secondPath
	isTargetIsFolder = isDir(secondPath)
	if isTargetIsFolder {
		file := path.Base(sourcePath)
		filename = secondPath + file
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalln("failed to create file %q, %v", filename, err)
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
		log.Fatalln("failed to download file, %v", err)
	}
	fmt.Printf("file downloaded, %d bytes\n", n)
}

// copy from local file to remove destination
func copyLocalToRemote() {
	isSourceIsFolder = isDir(firstPath)

	// parse target path
	args := strings.Split(secondPath, ":")
	isTargetIsFolder = strings.HasSuffix(secondPath, "/")
	// region:bucket:path
	if len(args) < 3 {
		log.Fatalln("S3 path must be in format region:bucketname:targetpath")
	}
	region = args[0]
	bucket = args[1]
	destPath = args[2]

	// check for region
	if !stringInSlice(region, predefinedRegions) {
		log.Fatalln("Invalid Region Name " + region)
	}

	// if source is folder, the target must be a folder
	if isSourceIsFolder && !isTargetIsFolder {
		log.Fatalln("When source path is folder - target path also need to be a folder (end with /)")
	}

	if isTargetIsFolder {
		prefix = destPath
	}

	// walk the files
	walker := make(fileWalk)
	go func() {
		// Gather the files to upload by walking the path recursively.
		if err := filepath.Walk(firstPath, walker.Walk); err != nil {
			log.Fatalln("Walk failed:", err)
		}
		close(walker)
	}()

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	uploader := s3manager.NewUploader(sess)

	// For each file found walking upload it to S3.
	for path := range walker {
		rel, err := filepath.Rel(firstPath, path)
		if err != nil {
			log.Fatalln("Unable to get relative path:", path, err)
		}
		file, err := os.Open(path)
		if err != nil {
			log.Println("Failed opening file", path, err)
			continue
		}
		defer file.Close()

		targetKey := destPath
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
}

type fileWalk chan string

func (f fileWalk) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		f <- path
	}
	return nil
}

func isDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

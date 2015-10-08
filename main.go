package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/codegangsta/cli"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	firstPath  string
	secondPath string
	prefix     string

	region   string
	bucket   string
	destPath string

	// true when source is a folder (so we need to walk on it, false when not)
	isSourceIsFolder bool

	// true when target is a folder (end with / e.g. south:bucket:/root/one/wise/, false when not - will be assumed as filename)
	isTargetIsFolder bool

	// by default all uploaded file will be `private` you need to specify public flag
	isPublic bool

	// verbose mode
	isVerbose bool
)

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
			Usage: "upload as Public ACL",
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
				fmt.Println("   us-east-1")
				fmt.Println("   us-west-1")
				fmt.Println("   us-west-2")
				fmt.Println("   eu-west-1")
				fmt.Println("   eu-central-1")
				fmt.Println("   ap-southeast-1")
				fmt.Println("   ap-southeast-2")
				fmt.Println("   ap-northeast-1")
				fmt.Println("   sa-east-1")
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

		isVerbose = c.Bool("verbose")
		isPublic = c.Bool("public")
		copyLocalToRemote()
	}

	app.Run(os.Args)
}

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

	// new aws s3 client
	config := aws.NewConfig().WithRegion(region)
	if isVerbose {
		config = config.WithLogLevel(aws.LogDebug)
	}
	s3client := s3.New(config)

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

		// check permission
		acl := aws.String("authenticated-read")
		if isPublic {
			acl = aws.String("public-read")
		}

		params := &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(targetKey),
			ACL:    acl,
			Body:   file,
		}
		result, err := s3client.PutObject(params)
		if err != nil {
			log.Fatalln("Failed to upload", path, err)
		}
		fmt.Println("    Uploaded with ETag: ", result.ETag)
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

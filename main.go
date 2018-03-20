package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/azureru/go-s3cp/utils"
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
	// {
	// 	Name:    "region",
	// 	Aliases: []string{"reg"},
	// 	Usage:   "List all Amazon Regions (as a reminder :P)",
	// 	Action: func(c *cli.Context) {
	// 		// since the SDK does not provide ways to enumerate this -
	// 		fmt.Println("List of Regions:")
	// 		for index := 0; index < len(predefinedRegions); index++ {
	// 			fmt.Printf("   %s\n", predefinedRegions[index])
	// 		}
	// 	},
	// },
	// {
	// 	Name:  "permission",
	// 	Usage: "List all Permissions",
	// 	Action: func(c *cli.Context) {
	// 		fmt.Println("List of Permissions:")
	// 		for index := 0; index < len(predefinedPermissions); index++ {
	// 			fmt.Printf("   %s\n", predefinedPermissions[index])
	// 		}
	// 	},
	// },
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
		// if !stringInSlice(permissionName, predefinedPermissions) {
		// 	log.Fatalln("Invalid Permission Name " + permissionName)
		// }
		fmt.Println("Permission: " + permissionName)

		if strings.Contains(firstPath, ":") {
			copyRemoteToLocal()
		} else {
			copyLocalToRemote()
		}
	}

	app.Run(os.Args)
}

// copyRemoteToLocal copy remote file to local location
func copyRemoteToLocal() {
	utils.DownloadS3(firstPath, secondPath)
}

// copyLocalToRemote copy local file(s) to remove location
func copyLocalToRemote() {
	utils.UploadS3(firstPath, secondPath, permissionName, storageClass)
}

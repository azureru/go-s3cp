package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/azureru/goccp/utils"
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
)

func main() {
	app := cli.NewApp()
	app.Name = "goccp"
	app.Author = "Erwin Saputra"
	app.Version = "0.0.2"
	app.Usage = "Utility to copy files from/to s3 and GCS"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, vv",
			Usage: "show Verbose Output",
		},
		cli.BoolFlag{
			Name:  "public, p",
			Usage: "upload as Public ACL, this will ignore setting on -permission flag",
		},
		cli.StringFlag{
			Name:  "storage, s",
			Usage: "specify the storage class",
		},
		cli.StringFlag{
			Name:  "permission, perm",
			Usage: "specify the Permission of the file",
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
				for index := 0; index < len(utils.S3Regions); index++ {
					fmt.Printf("   %s\n", utils.S3Regions[index])
				}
			},
		},
	}

	app.Action = func(c *cli.Context) {
		args := c.Args()
		if len(args) < 2 {
			fmt.Println("You need to put [from] and [to] path.")
			fmt.Println("EXAMPLE:")
			fmt.Println("   goccp ./file s3:regionname:bucket:path/path/filename")
			fmt.Println("   goccp ./file gs://bucket/path/filename")
			fmt.Println("   goccp gs://bucket/path/filename ./")
			return
		}
		firstPath = args[0]
		secondPath = args[1]

		// storage class
		utils.GlobalParamArg.StorageClass = "STANDARD"
		if c.String("storage") != "" {
			utils.GlobalParamArg.StorageClass = c.String("storage")
		}
		fmt.Println("Storage Class: " + utils.GlobalParamArg.StorageClass)

		// verbosity
		utils.GlobalParamArg.IsVerbose = c.Bool("verbose")

		// permission
		utils.GlobalParamArg.PermissionName = "private"
		if c.String("permission") != "" {
			utils.GlobalParamArg.PermissionName = c.String("permission")
		}

		utils.GlobalParamArg.IsPublic = c.Bool("public")
		if utils.GlobalParamArg.IsPublic {
			utils.GlobalParamArg.PermissionName = "public-read"
		}

		// when the first path is s3:blah:blah or gs://blah/blah
		// we identify it as copyFrom remote to local path
		isFirstRemote := strings.Contains(firstPath, ":")
		isLastRemote := strings.Contains(secondPath, ":")

		log.Println(isFirstRemote, isLastRemote)

		if isFirstRemote && !isLastRemote {
			copyRemoteToLocal()
		} else if !isFirstRemote && isLastRemote {
			copyLocalToRemote()
		} else {
			log.Fatalln("The tool does not support remote to remote copy")
		}
	}

	app.Run(os.Args)
}

// copyRemoteToLocal copy remote file to local location
func copyRemoteToLocal() {
	if utils.IsS3(firstPath) {
		utils.DownloadS3(firstPath, secondPath)
	} else if utils.IsGCS(firstPath) {
		utils.DownloadGCS(firstPath, secondPath)
	} else {
		log.Fatalln("Only support s3:bucket:path or gs://bucket/path as source remote location")
	}
}

// copyLocalToRemote copy local file(s) to remove location
func copyLocalToRemote() {
	if utils.IsS3(secondPath) {
		utils.UploadS3(firstPath, secondPath, utils.GlobalParamArg.PermissionName, utils.GlobalParamArg.StorageClass)
	} else if utils.IsGCS(secondPath) {
		utils.UploadGCS(firstPath, secondPath, utils.GlobalParamArg.PermissionName, utils.GlobalParamArg.StorageClass)
	} else {
		log.Fatalln("Only support s3:bucket:path or gs://bucket/path as target remote location")
	}
}

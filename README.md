# go-s3cp

An attempt to learn about `golang` by creating a tool to copy file `from` and `to` AWS S3 and Google GCS

The reason this exists is to have no-dependency binary that can easily copy-pasted inside containers or your own server.

An UPX-ed golang binary should be small enough for the purpose

## AWS Credentials
The credential will use Amazon SDK style of storing credential (use aws on your home, on your ENV var, or on your EC2 binded role - if you run it on EC2) - More info here `https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Credentials`

## GCS Credentials
This also will use Google SDK style of credentials - more info here
`https://cloud.google.com/docs/authentication/production`

```
    go-s3cp help

    go-s3cp ./file.txt us-east-1:bucketname:/file.txt

    go-s3cp -vv -perm public-read ./test ap-southeast-1:bucketname:/targetpath/
```

## TODO
- Refactor the provider to an interface :P (so we can support digical ocean's for example, or Webdav or FTP or other :|)
- Refactor the directory walker to an interface

- For now copy from remote source cannot walk through nested-folders
  since that will incur additional API calls for listing

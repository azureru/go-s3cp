# go-s3cp

An attempt to learn about `golang`

S3 Copy Command Line Tool (using golang)
Basically `spiritually` replicate Bradley Lucas's `https://aws.amazon.com/code/Amazon-S3/3124` functionality

The credential will use Amazon SDK style of storing credential (use aws on your home,
on your ENV var, or on your EC2 binded role - if you run it on EC2) - More info here `https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Credentials`

To copy

```
PUT
    s3cp local-file-path s3://zone/bucket/object[/]

    If object has a trailing slash it will be assumed to mean a directory
    and the local-file's filename will be appended to object.

GET
    s3cp s3://zone/bucket/object local-file

    If local-file is not present a filename from object will be used in
    the current directory.
```

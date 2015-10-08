# go-s3cp

An attempt to learn about `golang`

S3 Copy Command Line Tool (using golang)
Basically `spiritually` replicate Bradley Lucas's `https://aws.amazon.com/code/Amazon-S3/3124` functionality

The credential will use Amazon SDK style of storing credential (use aws on your home,
on your ENV var, or on your EC2 binded role - if you run it on EC2) - More info here `https://github.com/aws/aws-sdk-go/wiki/Getting-Started-Credentials`

```
    go-s3cp help

    go-s3cp ./file.txt us-east-1:bucketname:/file.txt

    go-s3cp -vv -perm public-read ./test ap-southeast-1:bucketname:/targetpath/
```

# todo

- Copy from s3 to local
- Generate signed url of s3 remote file to upload on other s/w

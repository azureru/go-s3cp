package utils

import (
	"fmt"
	"testing"
)

func TestBucketExtract(t *testing.T) {
	region, bucket, path, err := ShortS3Extract("s3:ap-southeast1:bucket:path")
	if err != nil {
		t.Error("Fail to parse proper s3 path")
	}
	fmt.Println(region)
	fmt.Println(bucket)
	fmt.Println(path)

	_, _, _, err = ShortS3Extract("s3:bucket:path")
	if err == nil {
		t.Error("Fail - Invalid s3 path must give error")
	}

	_, _, _, err = ShortS3Extract("s3:path")
	if err == nil {
		t.Error("Fail - Invalid s3 path must give error")
	}

	_, _, _, err = ShortS3Extract("whatever:region:bucket:path")
	if err == nil {
		t.Error("Fail - Invalid s3 path must give error")
	}
}

func TestIsS3(t *testing.T) {
	res := IsS3("s3:ap-southeast1:bucket:path")
	if res != true {
		t.Error("Fail - valid s3 return IsS3 false")
	}

	res = IsS3("whatever:ap-southeast1:bucket:path")
	if res == true {
		t.Error("Fail - invalid s3 return IsS3 true")
	}
}

package utils

import (
	"fmt"
	"testing"
)

func TestBucketExtractGcs(t *testing.T) {
	bucket, path, err := ShortGCSExtract("gs://bucket/path/google/file.zip")
	if err != nil {
		t.Error("Fail to parse proper GCS path")
	}
	fmt.Println(bucket)
	fmt.Println(path)

	_, _, err = ShortGCSExtract("gs://bucket")
	if err == nil {
		t.Error("Fail - Invalid GCS path must give error")
	}

	_, _, err = ShortGCSExtract("GCS:path")
	if err == nil {
		t.Error("Fail - Invalid GCS path must give error")
	}

	_, _, _, err = ShortS3Extract("whatever:region:bucket:path")
	if err == nil {
		t.Error("Fail - Invalid GCS path must give error")
	}
}

func TestIsGCS(t *testing.T) {
	res := IsGCS("gs://bucket/path.zip")
	if res != true {
		t.Error("Fail - valid GCS return IsGCS false")
	}

	res = IsGCS("whatever:ap-southeast1:bucket:path")
	if res == true {
		t.Error("Fail - invalid GCS return IsS3 true")
	}
}

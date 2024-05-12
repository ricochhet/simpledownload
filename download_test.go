package simpledownload_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ricochhet/simpledownload"
)

const testDownloadURL = "https://raw.githubusercontent.com/ricochhet/ricochhet/main/README.md"

func TestGenericDownload(t *testing.T) {
	t.Parallel()

	testMessenger := simpledownload.DownloadMessenger{
		StartDownload: func(fname string) {
			fmt.Printf("Test download: %s\n", fname)
		},
	}

	if bytes, err := simpledownload.Download(testDownloadURL); err != nil || len(bytes) == 0 {
		t.Fatal("download fail")
	}

	if bytes, err := simpledownload.DownloadWithContext(context.TODO(), testMessenger, testDownloadURL); err != nil || len(bytes) == 0 {
		t.Fatal("download fail")
	}
}

func TestFileDownload(t *testing.T) {
	t.Parallel()

	if err := simpledownload.File(testDownloadURL, "aaabbbccc", "README.md", "./.test/"); err != nil {
		t.Fatal(err)
	}

	if bytes, err := simpledownload.FileWithBytes(testDownloadURL, "aaabbbccc", "README.md", "./.test/"); err != nil || len(bytes) == 0 {
		t.Fatal("download fail")
	}
}

//nolint:lll // test only
func TestFileDownloadWithHash(t *testing.T) {
	t.Parallel()

	testMessenger := simpledownload.DownloadMessenger{
		StartDownload: func(fname string) {
			fmt.Printf("Test download: %s\n", fname)
		},
	}

	if err := simpledownload.FileWithContext(context.TODO(), testMessenger, testDownloadURL, "1c2b178a5fe8919c97c199cd1bea39075ef7260bed3794922a376896831f1516", "README.md", "./.test/", simpledownload.DefaultHashValidator); err != nil {
		t.Fatal(err)
	}

	if err := simpledownload.FileWithContext(context.TODO(), testMessenger, testDownloadURL, "", "README.md", "./.test/", simpledownload.DefaultHashValidator); err == nil {
		t.Fatal("empty hash has validated successfully")
	}

	if bytes, err := simpledownload.FileWithBytesAndContext(context.TODO(), testMessenger, testDownloadURL, "", "README.md", "./.test/", nil); err != nil || len(bytes) == 0 {
		t.Fatal("download fail")
	}
}

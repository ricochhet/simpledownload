package simpledownload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	errDownloadURLEmpty  = errors.New("download url is empty")
	errDownloadPathEmpty = errors.New("download path is empty")
	errDownloadNameEmpty = errors.New("download name is empty")
	errFileHashNoMatch   = errors.New("file hash does not match")
)

type DownloadMessenger struct {
	StartDownload func(string)
}

func DefaultDownloadMessenger() DownloadMessenger {
	return DownloadMessenger{
		StartDownload: func(fileName string) {
			fmt.Printf("%s ... DOWNLOADING\n", fileName)
		},
	}
}

func DefaultHashValidator(filePath, fileHash, fileName string) error {
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err == nil {
			hash := sha256.New()
			hash.Write(data)
			hashSum := hex.EncodeToString(hash.Sum(nil))

			if strings.ToLower(fileHash) == hashSum {
				fmt.Printf("%s ... OK\n", fileName)
				return nil
			}
		}
	}

	return errFileHashNoMatch
}

func File(url, fileHash, fileName, filePath string) error {
	return FileWithContext(context.TODO(), DefaultDownloadMessenger(), url, fileHash, fileName, filePath, nil)
}

func FileWithBytes(url, fileHash, fileName, filePath string) ([]byte, error) {
	if err := FileWithContext(context.TODO(), DefaultDownloadMessenger(), url, fileHash, fileName, filePath, nil); err != nil {
		return nil, err
	}

	return read(filePath, fileName)
}

//nolint:lll // wontfix
func FileWithBytesAndContext(ctx context.Context, state DownloadMessenger, url, fileHash, fileName, filePath string, hashValidator func(string, string, string) error) ([]byte, error) {
	if err := FileWithContext(ctx, state, url, fileHash, fileName, filePath, hashValidator); err != nil {
		return nil, err
	}

	return read(filePath, fileName)
}

//nolint:lll // wontfix
func FileWithContext(ctx context.Context, state DownloadMessenger, url, fileHash, fileName, filePath string, hashValidator func(string, string, string) error) error {
	if err := validateDownloadParams(url, filePath, fileName); err != nil {
		return err
	}

	fpath := filepath.Join(filePath, fileName)
	if err := os.MkdirAll(filepath.Dir(fpath), 0o700); err != nil {
		return err
	}

	if hashValidator != nil {
		if err := hashValidator(fpath, fileHash, fileName); err == nil {
			return nil
		}
	}

	state.StartDownload(fileName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	if hashValidator != nil {
		return write(resp, file, fileHash, fileName, false)
	}

	return write(resp, file, fileHash, fileName, true)
}

func validateDownloadParams(url, apath, name string) error {
	if url == "" {
		return errDownloadURLEmpty
	}

	if apath == "" {
		return errDownloadPathEmpty
	}

	if name == "" {
		return errDownloadNameEmpty
	}

	return nil
}

func read(filePath, fileName string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(filePath, fileName))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func write(resp *http.Response, flags *os.File, fileHash, fileName string, skipHashValidation bool) error {
	hash := sha256.New()
	buf := make([]byte, 1<<20) //nolint:mnd // 1 megabyte buffer

	for {
		index, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if index == 0 {
			break
		}

		if _, err := flags.Write(buf[:index]); err != nil {
			return err
		}

		if _, err := hash.Write(buf[:index]); err != nil {
			return err
		}
	}

	if skipHashValidation {
		return nil
	}

	hashSum := hex.EncodeToString(hash.Sum(nil))
	if strings.ToLower(fileHash) != hashSum {
		return fmt.Errorf("hash mismatch for %s", fileName) //nolint:err113 // required name prevents static error
	}

	return nil
}

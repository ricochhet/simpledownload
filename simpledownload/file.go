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
)

func File(url, checkHash, fileName, filePath string) ([]byte, error) {
	return FileWithContext(context.TODO(), url, checkHash, fileName, filePath)
}

func FileWithContext(ctx context.Context, url, checkHash, fileName, filePath string) ([]byte, error) {
	if err := validateDownloadParams(url, filePath, fileName, fileName); err != nil {
		return nil, err
	}

	fpath := filepath.Join(filePath, fileName)
	if err := os.MkdirAll(filepath.Dir(fpath), 0o700); err != nil {
		return nil, err
	}

	if _, err := os.Stat(fpath); err == nil {
		data, err := os.ReadFile(fpath)
		if err == nil {
			hash := sha256.New()
			hash.Write(data)
			hashSum := hex.EncodeToString(hash.Sum(nil))

			if strings.ToLower(checkHash) == hashSum {
				fmt.Printf("%s ... OK\n", fileName)
				return data, nil
			}
		}
	}

	fmt.Printf("%s ... DOWNLOADING\n", fileName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	file, err := os.Create(fpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return hashMatch(resp, file, fpath, checkHash, fileName)
}

func validateDownloadParams(url, apath, name, aname string) error {
	if url == "" {
		return errDownloadURLEmpty
	}

	if apath == "" {
		return errDownloadPathEmpty
	}

	if name == "" || aname == "" {
		return errDownloadNameEmpty
	}

	return nil
}

func hashMatch(resp *http.Response, flags *os.File, fpath, check, name string) ([]byte, error) {
	hash := sha256.New()
	buf := make([]byte, 1<<20) //nolint:mnd // 1 megabyte buffer

	for {
		index, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if index == 0 {
			break
		}

		if _, err := flags.Write(buf[:index]); err != nil {
			return nil, err
		}

		if _, err := hash.Write(buf[:index]); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	hashSum := hex.EncodeToString(hash.Sum(nil))
	if strings.ToLower(check) != hashSum {
		return nil, fmt.Errorf("hash mismatch for %s", name) //nolint:err113 // required name prevents static error
	}

	return data, nil
}

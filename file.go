/*
 * simpledownload
 * Copyright (C) 2024 simpledownload contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

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

func File(url, fileName, filePath string) error {
	return FileWithContext(context.TODO(), DefaultDownloadMessenger(), url, "", fileName, filePath, nil)
}

func FileValidated(url, fileHash, fileName, filePath string) error {
	return FileWithContext(context.TODO(), DefaultDownloadMessenger(), url, fileHash, fileName, filePath, DefaultHashValidator)
}

func FileWithBytes(url, fileName, filePath string) ([]byte, error) {
	if err := FileWithContext(context.TODO(), DefaultDownloadMessenger(), url, "", fileName, filePath, nil); err != nil {
		return nil, err
	}

	return read(filePath, fileName)
}

//nolint:lll // wontfix
func FileWithBytesValidated(url, fileHash, fileName, filePath string) ([]byte, error) {
	if err := FileWithContext(context.TODO(), DefaultDownloadMessenger(), url, fileHash, fileName, filePath, DefaultHashValidator); err != nil {
		return nil, err
	}

	return read(filePath, fileName)
}

//nolint:lll // wontfix
func FileWithContextAndBytes(ctx context.Context, state DownloadMessenger, url, fileHash, fileName, filePath string, hashValidator func(string, string, string) error) ([]byte, error) {
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

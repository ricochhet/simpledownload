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

package simpledownload_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ricochhet/simpledownload"
)

const testDownloadURL = "https://raw.githubusercontent.com/ricochhet/simpledownload/main/LICENSE"

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

	if err := simpledownload.File(testDownloadURL, "LICENSE", "./.test/"); err != nil {
		t.Fatal(err)
	}

	if bytes, err := simpledownload.FileWithBytes(testDownloadURL, "LICENSE", "./.test/"); err != nil || len(bytes) == 0 {
		t.Fatal("download fail")
	}
}

//nolint:lll // test only
func TestFileValidated(t *testing.T) {
	t.Parallel()

	if err := simpledownload.FileValidated(testDownloadURL, "aaabbbccc", "LICENSE", "./.test/"); err == nil {
		t.Fatal("download fail")
	}

	if bytes, err := simpledownload.FileWithBytesValidated(testDownloadURL, "aaabbbccc", "LICENSE", "./.test/"); err == nil || len(bytes) != 0 {
		t.Fatal("download fail")
	}

	if err := simpledownload.FileValidated(testDownloadURL, "8486a10c4393cee1c25392769ddd3b2d6c242d6ec7928e1414efff7dfb2f07ef", "LICENSE", "./.test/"); err != nil {
		t.Fatal(err)
	}

	if bytes, err := simpledownload.FileWithBytesValidated(testDownloadURL, "8486a10c4393cee1c25392769ddd3b2d6c242d6ec7928e1414efff7dfb2f07ef", "LICENSE", "./.test/"); err != nil || len(bytes) == 0 {
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

	if err := simpledownload.FileWithContext(context.TODO(), testMessenger, testDownloadURL, "8486a10c4393cee1c25392769ddd3b2d6c242d6ec7928e1414efff7dfb2f07ef", "LICENSE", "./.test/", simpledownload.DefaultHashValidator); err != nil {
		t.Fatal(err)
	}

	if err := simpledownload.FileWithContext(context.TODO(), testMessenger, testDownloadURL, "", "LICENSE", "./.test/", simpledownload.DefaultHashValidator); err == nil {
		t.Fatal("empty hash has validated successfully")
	}

	if bytes, err := simpledownload.FileWithContextAndBytes(context.TODO(), testMessenger, testDownloadURL, "", "LICENSE", "./.test/", nil); err != nil || len(bytes) == 0 {
		t.Fatal("download fail")
	}
}

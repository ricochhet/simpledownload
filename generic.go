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
	"io"
	"net/http"
)

func Download(url string) ([]byte, error) {
	return DownloadWithContext(context.TODO(), DefaultDownloadMessenger(), url)
}

func DownloadWithContext(ctx context.Context, messenger DownloadMessenger, url string) ([]byte, error) {
	if url == "" {
		return nil, errDownloadURLEmpty
	}

	messenger.StartDownload(url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

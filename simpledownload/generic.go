package simpledownload

import (
	"context"
	"io"
	"net/http"
)

func Download(url string) ([]byte, error) {
	return DownloadWithContext(context.TODO(), url)
}

func DownloadWithContext(ctx context.Context, url string) ([]byte, error) {
	if url == "" {
		return nil, errDownloadURLEmpty
	}

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

package opensearch

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOpensearch(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	url, err := container.GetHttpHostAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "http://") {
		t.Errorf("url: want prefix 'http://', got '%s'", url)
	}

	client := http.Client{}
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(request)
	if err != nil {
		t.Error(err)
	}
	t.Logf("code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: want=%d, got=%d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	t.Logf("body: %s", string(body))
}

func TestOpensearch_WithSecurityEnabled(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, WithSecurityEnabled())
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	url, err := container.GetHttpHostAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("url: want prefix 'https://', got '%s'", url)
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.SetBasicAuth(container.GetUserName(), container.GetPassword())

	resp, err := client.Do(request)
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: want=%d, got=%d", http.StatusOK, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	t.Logf("body: %s", string(body))
}

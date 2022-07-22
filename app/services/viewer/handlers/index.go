package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	index, err := os.Open("app/services/viewer/assets/views/index.html")
	if err != nil {
		return fmt.Errorf("open index page: %w", err)
	}
	defer index.Close()

	io.Copy(w, index)

	return nil
}

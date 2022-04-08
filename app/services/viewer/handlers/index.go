package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
)

type indexGroup struct {
	tmpl *template.Template
}

func newIndex() (indexGroup, error) {
	index, err := os.Open("app/services/viewer/assets/views/index.tmpl")
	if err != nil {
		return indexGroup{}, fmt.Errorf("open index page: %w", err)
	}
	defer index.Close()

	rawTmpl, err := io.ReadAll(index)
	if err != nil {
		return indexGroup{}, fmt.Errorf("reading index page: %w", err)
	}

	tmpl := template.New("index")
	if _, err := tmpl.Parse(string(rawTmpl)); err != nil {
		return indexGroup{}, fmt.Errorf("creating template: %w", err)
	}

	ig := indexGroup{
		tmpl: tmpl,
	}

	return ig, nil
}

func (ig *indexGroup) handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var markup bytes.Buffer
	vars := map[string]any{
		"variable": "testing",
	}

	if err := ig.tmpl.Execute(&markup, vars); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	io.Copy(w, &markup)

	return nil
}

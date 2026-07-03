package http

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/wfrs-dev/sic/internal/fn"
	"github.com/wfrs-dev/sic/internal/model"
)

const baseURL = "https://start.spring.io"

func FirstRequest() (gjson.Result, error) {
	req, err := http.NewRequest("GET", baseURL+"/", nil)
	if err != nil {
		return gjson.Result{}, err
	}
	req.Header.Set("Accept", "application/vnd.initializr.v2.1+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return gjson.Result{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(body), nil
}

func CreateProject(project model.ProjectRequest, target string) error {
	data := url.Values{}
	data.Set("type", project.Project)
	data.Set("language", project.Language)
	data.Set("bootVersion", project.SpringBoot)
	data.Set("groupId", project.Group)
	data.Set("artifactId", project.Artifact)
	data.Set("packageName", project.PackageName)
	data.Set("packaging", project.Packaging)
	data.Set("javaVersion", project.JavaVersion)
	data.Set("name", project.Name)
	data.Set("description", project.Description)
	data.Set("dependencies", project.Dependencies)

	slog.Debug("Sending request", slog.String("url", baseURL+target), slog.String("data", data.Encode()))
	resp, err := http.PostForm(baseURL+target, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if strings.Contains(mediaType, "zip") {
		baseDir := fn.Slugify(project.Name)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return fmt.Errorf("error creating project directory: %w", err)
		}
		slog.Debug("Creating project directory", slog.String("dir", baseDir))

		return extractZip(body, baseDir)
	} else if strings.Contains(mediaType, "json") {
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}

		if v, ok := data["status"]; ok && v.(int) == 500 {
			return fmt.Errorf("error creating project: %s", data["message"])
		}

		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errors.New("error formatting response")
		}

		return errors.New(string(out))
	}

	return os.WriteFile(filepath.Base(target), body, 0644)
}

func extractZip(data []byte, baseDir string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		path := filepath.Join(baseDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		os.MkdirAll(filepath.Dir(path), 0755)

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

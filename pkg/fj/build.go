package fj

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"k8s.io/klog/v2"
)

//go:embed assets/stream.tmpl
var streamTmpl string

//go:embed assets/stream.css
var streamCSS string

func Build(inDir string, outDir string) error {
	klog.Infof("build: %s -> %s", inDir, outDir)

	is, err := Find(inDir)
	klog.Infof("images: %+v", is)

	for _, i := range is {
		i.Thumbnails, err = thumbnails(*i, outDir)
		if err != nil {
			return fmt.Errorf("thumbnails: %v", err)
		}

		i.ThumbPath = i.Thumbnails["512x"].RelPath

		if i.ThumbPath == "" {
			return fmt.Errorf("unable to find thumb for %+v", i)
		}

		klog.Infof("thumbpath: %s", i.ThumbPath)
	}

	html, err := renderStream("fj stream", is)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(outDir, "index.html"), []byte(html), 0644)
	return err
}

func renderStream(title string, is []*Image) (string, error) {
	funcMap := template.FuncMap{
		"Odd": func(i int) bool {
			if i%2 == 1 {
				return true
			}
			return false
		},
	}
	tmpl, err := template.New("stream").Funcs(funcMap).Parse(streamTmpl)
	if err != nil {
		return "", fmt.Errorf("parse: %v", err)
	}

	data := struct {
		Title      string
		Stylesheet template.CSS
		Images     []*Image
	}{
		Title:      title,
		Stylesheet: template.CSS(streamCSS),
		Images:     is,
	}

	var tpl bytes.Buffer
	if err = tmpl.Execute(&tpl, data); err != nil {
		return "", fmt.Errorf("execute: %w", err)
	}

	out := tpl.String()
	return out, nil
}
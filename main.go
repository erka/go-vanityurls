package main

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	var configPath, outputDir string
	flag.StringVar(&configPath, "config", "vanity.yaml", "path to vanity config file")
	flag.StringVar(&outputDir, "output", "public", "output directory for generated files")

	flag.Parse()

	input, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("failed to read the config:", err)
	}
	if err := generate(input, outputDir); err != nil {
		log.Fatal("failed to generate:", err)
	}
}

type pathConfig struct {
	path    string
	repo    string
	display string
	vcs     string
	subdir  string
}

type config struct {
	Host  string `yaml:"host,omitempty"`
	Paths map[string]struct {
		Repo    string `yaml:"repo,omitempty"`
		Display string `yaml:"display,omitempty"`
		VCS     string `yaml:"vcs,omitempty"`
		Subdir  string `yaml:"subdir,omitempty"`
	} `yaml:"paths,omitempty"`
}

func generate(input []byte, outputDir string) error {
	var cfg config
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		return err
	}

	paths := make([]pathConfig, 0, len(cfg.Paths))
	for path, e := range cfg.Paths {
		p := pathConfig{
			path:    strings.TrimSuffix(path, "/"),
			repo:    e.Repo,
			display: e.Display,
			vcs:     e.VCS,
			subdir:  e.Subdir,
		}
		p.vcs = cmp.Or(p.vcs, "git")
		p.display = cmp.Or(p.display, fmt.Sprintf("%v %v/tree/main{/dir} %v/blob/main{/dir}/{file}#L{line}", e.Repo, e.Repo, e.Repo))
		paths = append(paths, p)
	}
	slices.SortFunc(paths, func(a pathConfig, b pathConfig) int {
		return cmp.Compare(a.path, b.path)
	})

	return generateFiles(cfg.Host, paths, outputDir)
}

func generateFiles(host string, paths []pathConfig, targetDir string) error {
	err := os.MkdirAll(targetDir, 0o700)
	if err != nil {
		return err
	}
	handlers := make([]string, len(paths))
	for i, h := range paths {
		handlers[i] = host + h.path
	}
	b := bytes.Buffer{}
	if err := indexTmpl.Execute(&b, struct {
		Host     string
		Handlers []string
	}{
		Host:     host,
		Handlers: handlers,
	}); err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(targetDir, "index.html"), b.Bytes(), 0o600)
	if err != nil {
		return err
	}

	for _, p := range paths {
		b.Reset()
		if subpath := filepath.Dir(p.path); subpath != "." {
			subpath := strings.TrimPrefix(subpath, "/")
			err = os.MkdirAll(filepath.Join(targetDir, subpath), 0o700)
			if err != nil {
				return err
			}
		}
		if err := vanityTmpl.Execute(&b, struct {
			Import  string
			Subpath string
			Repo    string
			Display string
			VCS     string
			Subdir  string
		}{
			Import:  host + p.path,
			Subpath: "",
			Repo:    p.repo,
			Display: p.display,
			VCS:     p.vcs,
			Subdir:  p.subdir,
		}); err != nil {
			return err
		}

		filename := filepath.Join(targetDir, strings.TrimPrefix(p.path, "/")+".html")
		err := os.WriteFile(filename, b.Bytes(), 0o600)
		if err != nil {
			return err
		}
	}
	return nil
}

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>go.openfeature.dev</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="description" content="Go vanity import paths for OpenFeature Go SDK and community modules, including providers and hooks.">
  <style>
		:root{--bg:#ffffff;--fg:#121416;--accent:#5d5dff}
		@media (prefers-color-scheme:dark){:root{--bg:#1b1b1c;--fg:#e3e3e3;--accent:#8b8bff}}
		html,body{margin:0;padding:0;background:var(--bg);color:var(--fg);line-height:1.1}main{margin:1rem auto;padding:0 1.5rem;max-width:600px}h1{font-size:1.4rem;margin:0 0 2rem;font-weight:600;color:var(--fg);text-align:center}ul{list-style:none;padding:0;margin:0}a{color:var(--fg);display:block;padding:.1rem .1rem;border-radius:6px;text-decoration:none;font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.92rem;transition:background 0.15s ease}a:hover{color:var(--accent)}.logo-img{display:block;margin:1rem auto}
  </style>
</head>
<body>
	<picture>
		<source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/open-feature/community/0e23508c163a6a1ac8c0ced3e4bd78faafe627c7/assets/logo/horizontal/white/openfeature-horizontal-white.svg">
		<img src="https://raw.githubusercontent.com/open-feature/community/0e23508c163a6a1ac8c0ced3e4bd78faafe627c7/assets/logo/horizontal/black/openfeature-horizontal-black.svg"
			alt="OpenFeature"
			width="497"
			height="69"
			class="logo-img">
	</picture>
	<h1>{{.Host}}</h1>
	<main>
	<ul>
	{{range .Handlers}}<li><a href="https://pkg.go.dev/{{.}}">{{.}}</a></li>{{end}}
	</ul>
	</main>
</body>
</html>
`))

var vanityTmpl = template.Must(template.New("vanity").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Import}} {{.VCS}} {{.Repo}}{{if .Subdir}} {{.Subdir}}{{end}}">
<meta name="go-source" content="{{.Import}} {{.Display}}">
<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/{{.Import}}/{{.Subpath}}">
</head>
<body>
Nothing to see here; <a href="https://pkg.go.dev/{{.Import}}/{{.Subpath}}">see the package on pkg.go.dev</a>.
</body>
</html>`))

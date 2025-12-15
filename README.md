# OpenFeature Go Vanity URLs Registry

This repository contains the vanity URL registry for OpenFeature Go packages. It uses a tool to generate vanity URLs for Go packages by converting a YAML configuration into static HTML files that serve as Go import redirects, with the generated files published to GitHub Pages in Go's expected format.

## Configuration

This repository uses `vanity.yaml` to define all official supported OpenFeature Go packages. The configuration file maps import paths to their repository locations.

### vanity.yaml Structure

```yaml
host: go.openfeature.dev

paths:
  /openfeature/v2:
    repo: https://github.com/open-feature/go-sdk

  /contrib/providers/aws-ssm/v2:
    repo: https://github.com/open-feature/go-sdk-contrib
    subdir: providers/aws-ssm
```

**Configuration Options:**

- `host`: Base domain for the vanity URLs
- `paths`: Map of import paths to repository configurations
  - `repo`: Repository URL (required)
  - `subdir`: Subdirectory within the repository containing the Go module (optional)
  - `vcs`: Version control system (default: `git`)
  - `display`: Custom go-source display format (optional)

### Manual Generation

To generate the vanity URL files locally:

```bash
go run main.go
```

This generates static HTML files in the `public` directory. The generated files follow Go's standard import redirect format with proper meta tags.

## Publishing with GitHub Actions & GitHub Pages

This repository includes a GitHub Actions workflow (`.github/workflows/vanity-gh-pages.yml`) that automatically publishes changes to GitHub Pages.

**Workflow Details:**

- **Trigger**: Automatically runs on every push to the `main` branch
- **Manual Trigger**: Can also be triggered manually via `workflow_dispatch`
- **Steps**:
  1. Checks out the repository code
  2. Sets up Go environment
  3. Generates vanity URL HTML files by running the tool
  4. Uploads the generated files as artifacts
  5. Deploys the artifacts to GitHub Pages

The generated HTML files are automatically deployed to GitHub Pages. The site is published from the `gh-pages` branch and is accessible at `https://go.openfeature.dev`.

The HTML files contain Go import metadata that redirect to pkg.go.dev.

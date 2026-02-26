package compliance

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

const (
	trivyChartName = "trivy-operator"
	trivyRepoURL   = "https://aquasecurity.github.io/helm-charts"
	trivyVersion   = "0.32.0"

	// Adjust if Aqua ever changes the release/tag naming.
	trivyHelmReleasesBase = "https://github.com/aquasecurity/helm-charts/releases/download"
)

func NewTrivyChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	values := map[string]any{
		"trivy-operator": map[string]any{
			"enabled": true,
			"serviceMonitor": map[string]any{
				"enabled": true,
			},
		},
		"operator": map[string]any{
			"replicas":                1,
			"scanJobsConcurrentLimit": 3,
			"scanJobsRetryDelay":      "30s",
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"trivyOperator": map[string]any{
			"scanJobTimeout": "5m",
		},
		"compliance": map[string]any{
			"failedChecksOnly": false,
		},
		"rbac": map[string]any{
			"create": true,
		},
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "trivy-operator",
		},
	}

	if os.Getenv("TRIVY_OPERATOR_SKIP_CRDS") != "true" {
		includeTrivyCRDs(chart)
	}

	cdk8s.NewHelm(chart, jsii.String("trivy-release"), &cdk8s.HelmProps{
		Chart:       jsii.String(trivyChartName),
		Repo:        jsii.String(trivyRepoURL),
		Version:     jsii.String(trivyVersion),
		ReleaseName: jsii.String("trivy-operator"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}

func includeTrivyCRDs(chart cdk8s.Chart) {
	tempDir, err := os.MkdirTemp("", "trivy-crds")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp dir for trivy CRDs: %v", err))
	}
	defer os.RemoveAll(tempDir)

	chartURL := fmt.Sprintf(
		"%s/%s-%s/%s-%s.tgz",
		trivyHelmReleasesBase,
		trivyChartName, trivyVersion,
		trivyChartName, trivyVersion,
	)

	resp, err := http.Get(chartURL)
	if err != nil {
		panic(fmt.Sprintf("failed to download trivy chart from %s: %v", chartURL, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("failed to download trivy chart from %s: status %s", chartURL, resp.Status))
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to create gzip reader for trivy chart: %v", err))
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("failed to read tar header from trivy chart: %v", err))
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Match CRDs inside the chart's crds/ directory
		if !strings.Contains(header.Name, "crds/") || !strings.HasSuffix(header.Name, ".yaml") {
			continue
		}

		targetPath := filepath.Join(tempDir, filepath.Base(header.Name))
		outFile, err := os.Create(targetPath)
		if err != nil {
			panic(fmt.Sprintf("failed to create CRD file %s: %v", targetPath, err))
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			panic(fmt.Sprintf("failed to write CRD file %s: %v", targetPath, err))
		}
		outFile.Close()

		id := fmt.Sprintf("crd-%s", strings.TrimSuffix(filepath.Base(header.Name), ".yaml"))

		cdk8s.NewInclude(chart, jsii.String(id), &cdk8s.IncludeProps{
			Url: jsii.String(targetPath),
		})
	}
}

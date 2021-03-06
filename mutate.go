package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cleanName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

func mutate(request v1beta1.AdmissionRequest) (v1beta1.AdmissionResponse, error) {
	response := v1beta1.AdmissionResponse{}

	// Default response
	response.Allowed = true
	response.UID = request.UID

	// Decode the pod object
	var err error
	pod := v1.Pod{}
	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		return response, fmt.Errorf("unable to decode Pod %w", err)
	}

	log.Printf("Check pod for notebook or workflow %s/%s", pod.Namespace, pod.Name)

	shouldInject := false
	// Inject Minio credentials into notebook pods (condition: has notebook-name label)
	if _, ok := pod.ObjectMeta.Labels["notebook-name"]; ok {
		log.Printf("Found notebook name for %s/%s", pod.Namespace, pod.Name)
		shouldInject = true
	}

	// Inject Minio credentials into argo workflow pods (condition: has workflows.argoproj.io/workflow label)
	if _, ok := pod.ObjectMeta.Labels["workflows.argoproj.io/workflow"]; ok {
		log.Printf("Found argo workflow name for %s/%s", pod.Namespace, pod.Name)
		shouldInject = true
	}

	// Inject Minio credentials into pod requesting credentials (condition: has add-default-minio-creds annotation)
	if _, ok := pod.ObjectMeta.Annotations["data.statcan.gc.ca/inject-minio-creds"]; ok {
		log.Printf("Found minio credential annotation on %s/%s", pod.Namespace, pod.Name)
		shouldInject = true
	}


	if shouldInject {
		patch := v1beta1.PatchTypeJSONPatch
		response.PatchType = &patch

		response.AuditAnnotations = map[string]string{
			"minio-admission-controller": "Added minio credentials",
		}

		roleName := cleanName("profile-" + pod.Namespace)

		patches := []map[string]interface{}{
			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject",
				"value": "true",
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-pre-populate",
				"value": "false",
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1role",
				"value": roleName,
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-minimal-tenant1",
				"value": "minio_minimal_tenant1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-minimal-tenant1",
				"value": fmt.Sprintf(`
{{- with secret "minio_minimal_tenant1/keys/%s" }}
export MINIO_URL="http://minimal-tenant1-minio.minio:9000"
export MINIO_ACCESS_KEY="{{ .Data.accessKeyId }}"
export MINIO_SECRET_KEY="{{ .Data.secretAccessKey }}"
export AWS_ACCESS_KEY_ID="{{ .Data.accessKeyId }}"
export AWS_SECRET_ACCESS_KEY="{{ .Data.secretAccessKey }}"
{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-minimal-tenant1.json",
				"value": "minio_minimal_tenant1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-minimal-tenant1.json",
				"value": fmt.Sprintf(`
{{- with secret "minio_minimal_tenant1/keys/%s" }}
{"MINIO_URL":"http://minimal-tenant1-minio.minio:9000","MINIO_ACCESS_KEY":"{{ .Data.accessKeyId }}","MINIO_SECRET_KEY":"{{ .Data.secretAccessKey }}","AWS_ACCESS_KEY_ID":"{{ .Data.accessKeyId }}","AWS_SECRET_ACCESS_KEY":"{{ .Data.secretAccessKey }}"}
{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-pachyderm-tenant1",
				"value": "minio_pachyderm_tenant1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-pachyderm-tenant1",
				"value": fmt.Sprintf(`
{{- with secret "minio_pachyderm_tenant1/keys/%s" }}
export MINIO_URL="http://pachyderm-tenant1-minio.minio:9000"
export MINIO_ACCESS_KEY="{{ .Data.accessKeyId }}"
export MINIO_SECRET_KEY="{{ .Data.secretAccessKey }}"
export AWS_ACCESS_KEY_ID="{{ .Data.accessKeyId }}"
export AWS_SECRET_ACCESS_KEY="{{ .Data.secretAccessKey }}"
{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-pachyderm-tenant1.json",
				"value": "minio_pachyderm_tenant1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-pachyderm-tenant1.json",
				"value": fmt.Sprintf(`
{{- with secret "minio_pachyderm_tenant1/keys/%s" }}
{"MINIO_URL":"http://pachyderm-tenant1-minio.minio:9000","MINIO_ACCESS_KEY":"{{ .Data.accessKeyId }}","MINIO_SECRET_KEY":"{{ .Data.secretAccessKey }}","AWS_ACCESS_KEY_ID":"{{ .Data.accessKeyId }}","AWS_SECRET_ACCESS_KEY":"{{ .Data.secretAccessKey }}"}
{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-premium-tenant1",
				"value": "minio_premium_tenant1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-premium-tenant1",
				"value": fmt.Sprintf(`
{{- with secret "minio_premium_tenant1/keys/%s" }}
export MINIO_URL="http://premium-tenant1-minio.minio:9000"
export MINIO_ACCESS_KEY="{{ .Data.accessKeyId }}"
export MINIO_SECRET_KEY="{{ .Data.secretAccessKey }}"
export AWS_ACCESS_KEY_ID="{{ .Data.accessKeyId }}"
export AWS_SECRET_ACCESS_KEY="{{ .Data.secretAccessKey }}"
{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-premium-tenant1.json",
				"value": "minio_premium_tenant1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-premium-tenant1.json",
				"value": fmt.Sprintf(`
{{- with secret "minio_premium_tenant1/keys/%s" }}
{"MINIO_URL":"http://premium-tenant1-minio.minio:9000","MINIO_ACCESS_KEY":"{{ .Data.accessKeyId }}","MINIO_SECRET_KEY":"{{ .Data.secretAccessKey }}","AWS_ACCESS_KEY_ID":"{{ .Data.accessKeyId }}","AWS_SECRET_ACCESS_KEY":"{{ .Data.secretAccessKey }}"}
{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-premium-tenant-1",
				"value": "minio_premium_tenant_1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-premium-tenant-1",
				"value": fmt.Sprintf(`
			{{- with secret "minio_premium_tenant_1/keys/%s" }}
			export MINIO_URL="http://minio.minio-premium-tenant-1"
			export MINIO_ACCESS_KEY="{{ .Data.accessKeyId }}"
			export MINIO_SECRET_KEY="{{ .Data.secretAccessKey }}"
			export AWS_ACCESS_KEY_ID="{{ .Data.accessKeyId }}"
			export AWS_SECRET_ACCESS_KEY="{{ .Data.secretAccessKey }}"
			{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-premium-tenant-1.json",
				"value": "minio_premium_tenant_1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-premium-tenant-1.json",
				"value": fmt.Sprintf(`
			{{- with secret "minio_premium_tenant_1/keys/%s" }}
			{"MINIO_URL":"http://minio.minio-premium-tenant-1","MINIO_ACCESS_KEY":"{{ .Data.accessKeyId }}","MINIO_SECRET_KEY":"{{ .Data.secretAccessKey }}","AWS_ACCESS_KEY_ID":"{{ .Data.accessKeyId }}","AWS_SECRET_ACCESS_KEY":"{{ .Data.secretAccessKey }}"}
			{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-standard-tenant-1",
				"value": "minio_standard_tenant_1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-standard-tenant-1",
				"value": fmt.Sprintf(`
			{{- with secret "minio_standard_tenant_1/keys/%s" }}
			export MINIO_URL="http://minio.minio-standard-tenant-1"
			export MINIO_ACCESS_KEY="{{ .Data.accessKeyId }}"
			export MINIO_SECRET_KEY="{{ .Data.secretAccessKey }}"
			export AWS_ACCESS_KEY_ID="{{ .Data.accessKeyId }}"
			export AWS_SECRET_ACCESS_KEY="{{ .Data.secretAccessKey }}"
			{{- end }}
						`, roleName),
			},

			{
				"op":    "add",
				"path":  "/metadata/annotations/vault.hashicorp.com~1agent-inject-secret-minio-standard-tenant-1.json",
				"value": "minio_standard_tenant_1/keys/" + roleName,
			},

			{
				"op":   "add",
				"path": "/metadata/annotations/vault.hashicorp.com~1agent-inject-template-minio-standard-tenant-1.json",
				"value": fmt.Sprintf(`
			{{- with secret "minio_standard_tenant_1/keys/%s" }}
			{"MINIO_URL":"http://minio.minio-standard-tenant-1","MINIO_ACCESS_KEY":"{{ .Data.accessKeyId }}","MINIO_SECRET_KEY":"{{ .Data.secretAccessKey }}","AWS_ACCESS_KEY_ID":"{{ .Data.accessKeyId }}","AWS_SECRET_ACCESS_KEY":"{{ .Data.secretAccessKey }}"}
			{{- end }}
						`, roleName),
			},
		}
		response.Patch, err = json.Marshal(patches)
		if err != nil {
			return response, err
		}

		response.Result = &metav1.Status{
			Status: metav1.StatusSuccess,
		}
	} else {
		log.Printf("Notebook name not found for %s/%s", pod.Namespace, pod.Name)
	}

	return response, nil
}

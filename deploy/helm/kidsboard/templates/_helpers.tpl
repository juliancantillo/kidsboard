{{/*
Expand the chart name (override allowed via .Values.nameOverride).
*/}}
{{- define "kidsboard.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Fully-qualified app name. Used for resource names. Truncated to 63 chars
(DNS label limit) and the release name is prepended unless overridden.
*/}}
{{- define "kidsboard.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Chart label (chart name + version).
*/}}
{{- define "kidsboard.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels — applied to every object the chart creates.
*/}}
{{- define "kidsboard.labels" -}}
helm.sh/chart: {{ include "kidsboard.chart" . }}
{{ include "kidsboard.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kidsboard
{{- end -}}

{{/*
Selector labels — used by Services and the StatefulSet to find Pods.
Must be a stable subset of common labels (no version-bound entries here).
*/}}
{{- define "kidsboard.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kidsboard.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Service account name. Returns the override or the default `<fullname>` if the
chart is configured to create one; otherwise falls back to "default".
*/}}
{{- define "kidsboard.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- include "kidsboard.fullname" . -}}
{{- else -}}
{{- "default" -}}
{{- end -}}
{{- end -}}

{{/*
Image reference. Tag defaults to .Chart.AppVersion when values.image.tag is empty.
*/}}
{{- define "kidsboard.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}

{{- define "life-is-hard.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "life-is-hard.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "life-is-hard.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "life-is-hard.chart" -}}
{{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}

{{- define "life-is-hard.labels" -}}
helm.sh/chart: {{ include "life-is-hard.chart" . }}
app.kubernetes.io/name: {{ include "life-is-hard.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "life-is-hard.selectorLabels" -}}
app.kubernetes.io/name: {{ include "life-is-hard.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "life-is-hard.postgresql.fullname" -}}
{{ printf "%s-postgresql" .Release.Name }}
{{- end -}}

{{- define "life-is-hard.redis.fullname" -}}
{{ printf "%s-redis" .Release.Name }}
{{- end -}}

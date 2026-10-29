{{- define "gfw.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "gfw.fullname" -}}
{{- if .Values.fullnameOverride }}{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}{{ else }}{{ printf "%s-%s" .Release.Name (include "gfw.name" .) | trunc 63 | trimSuffix "-" }}{{ end }}
{{- end }}

{{- define "gfw.chart" -}}{{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}{{- end }}

{{- define "gfw.labels" -}}
helm.sh/chart: {{ include "gfw.chart" . }}
{{ include "gfw.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

{{- define "gfw.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gfw.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "gfw.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}{{ default (include "gfw.fullname" .) .Values.serviceAccount.name }}{{ else }}{{ default "default" .Values.serviceAccount.name }}{{ end }}
{{- end }}

{{- define "gfw.secretName" -}}{{ default (include "gfw.fullname" .) .Values.secrets.existingSecret }}{{- end }}

{{- define "gfw.image" -}}
{{- printf "%s:%s" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) -}}
{{- end }}

{{- define "gfw.secretEnv" -}}
- name: GFW_DATABASE_URL
  valueFrom:
    secretKeyRef:
      name: {{ include "gfw.secretName" . }}
      key: {{ .Values.secrets.keys.databaseURL }}
- name: GFW_OIDC_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "gfw.secretName" . }}
      key: {{ .Values.secrets.keys.oidcClientSecret }}
- name: GFW_SESSION_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "gfw.secretName" . }}
      key: {{ .Values.secrets.keys.sessionSecret }}
- name: GFW_ADMIN_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ include "gfw.secretName" . }}
      key: {{ .Values.secrets.keys.adminToken }}
{{- end }}

{{- define "gfw.volumeMounts" -}}
- name: config
  mountPath: /etc/gfw/config.yaml
  subPath: config.yaml
  readOnly: true
- name: tmp
  mountPath: /tmp
{{- end }}

{{- define "gfw.volumes" -}}
- name: config
  configMap:
    name: {{ include "gfw.fullname" . }}
    defaultMode: 0440
- name: tmp
  emptyDir:
    sizeLimit: {{ .Values.tmp.sizeLimit }}
{{- end }}


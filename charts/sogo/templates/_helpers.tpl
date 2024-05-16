{{/*
Expand the name of the chart.
*/}}
{{- define "sogo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "sogo.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "sogo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create image name that is used in the deployment
*/}}
{{- define "sogo.image" -}}
{{- if .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository .Values.image.tag -}}
{{- else -}}
{{- if .Values.image.revision -}}
{{- printf "%s:%s-%s" .Values.image.repository .Chart.AppVersion .Values.image.revision -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.repository .Chart.AppVersion -}}
{{- end -}}
{{- end -}}
{{- end -}}


{{/*
Create DB URL paths
*/}}
{{- define "sogo.db.baseUrl" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "postgresql://%s:%s@%s:5432/%s" (or .Values.postgresql.global.postgresql.auth.username .Values.postgresql.auth.username) (or .Values.postgresql.global.postgresql.auth.password .Values.postgresql.auth.password) (template "postgresql.v1.primary.fullname" .Subcharts.postgresql) (or .Values.postgresql.global.postgresql.auth.database .Values.postgresql.auth.database) -}}
{{- else if .Values.mariadb.enabled -}}
{{- printf "mysql://%s:%s@%s:3306/%s" .Values.mariadb.auth.username .Values.mariadb.auth.password (template "mariadb.primary.fullname" .Subcharts.mariadb) .Values.mariadb.auth.database -}}
{{- end -}}
{{- end -}}

{{- define "sogo.db.configs" -}}
SOGoProfileURL: {{ printf "%s/sogo_user_profile" (include "sogo.db.baseUrl" .) }}
OCSFolderInfoURL: {{ printf "%s/sogo_folder_info" (include "sogo.db.baseUrl" .) }}
OCSSessionsFolderURL: {{ printf "%s/sogo_sessions_folder" (include "sogo.db.baseUrl" .) }}
OCSAdminURL: {{ printf "%s/sogo_admin" (include "sogo.db.baseUrl" .) }}
{{- end -}}

{{- define "sogo.memcached.configs" -}}
SOGoMemcachedHost: {{ template "common.names.fullname" .Subcharts.memcached }}
{{- end -}}

{{- $parts := split "://" (include "sogo.db.baseUrl" .) -}}
{{- $db_type := index $parts 0 -}}
{{- $remaining := index $parts 1 -}}

{{- define "sogo.db.type" -}}
{{- printf "%s" $db_type -}}
{{- end -}}

{{- define "sogo.ingress.apiVersion" -}}
{{- if semverCompare "<1.14-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "extensions/v1beta1" -}}
{{- else if semverCompare "<1.19-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "networking.k8s.io/v1beta1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1" -}}
{{- end }}
{{- end -}}


{{/*
Create volume mounts for the sogo container.
*/}}
{{- define "sogo.volumeMounts" -}}
{{- if .Values.sogo.extraVolumeMounts }}
{{ toYaml .Values.sogo.extraVolumeMounts }}
{{- end }}
{{- end -}}

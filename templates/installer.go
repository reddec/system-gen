package templates

import "text/template"

var ProjectInstallerTemplate = template.Must(template.New("").Parse(`#!/usr/bin/env bash
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root" 1>&2
   exit 1
fi
cd "$( dirname "${BASH_SOURCE[0]}")"

mkdir -p /etc/systemd/system
cp *.service /etc/systemd/system/
cp *.timer /etc/systemd/system/ 

systemctl enable {{.Slug}}
{{- range .Services}}
systemctl enable {{.Slug}}
{{- end}}
{{- range .OneShots}}
systemctl enable {{.Slug}}
{{- end}}
{{- range .Timers}}
systemctl enable {{.Slug}}.timer
{{- end}}

{{- range .OneShots}}
systemctl reset-failed {{.Slug}}
{{- end}}
{{- range .Services}}
systemctl reset-failed {{.Slug}}
{{- end}}
systemctl reset-failed {{.Slug}}
systemctl daemon-reload

echo "Start whole group:"
echo ""
echo "    systemctl start {{.Slug}}"
echo ""
echo "See running sub-units:"
echo ""
echo "    systemctl list-unit-files '{{.Slug}}*'"
echo ""
`))

var ProjectUnInstallerTemplate = template.Must(template.New("").Parse(`#!/usr/bin/env bash
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root" 1>&2
   exit 1
fi
cd "$( dirname "${BASH_SOURCE[0]}")"

systemctl stop {{.Slug}}
{{- range .Services}}
systemctl stop {{.Slug}}
{{- end}}
{{- range .OneShots}}
systemctl stop {{.Slug}}
{{- end}}
{{- range .Timers}}
systemctl stop {{.Slug}}.timer
{{- end}}

{{- range .Services}}
systemctl disable {{.Slug}}
{{- end}}
{{- range .OneShots}}
systemctl disable {{.Slug}}
{{- end}}
{{- range .Timers}}
systemctl disable {{.Slug}}.timer
{{- end}}

systemctl disable {{.Slug}}

{{- range .Services}}
rm /etc/systemd/system/{{.Slug}}.service
{{- end}}
{{- range .Timers}}
rm /etc/systemd/system/{{.Slug}}.timer
{{- end}}
rm /etc/systemd/system/{{.Slug}}.service

systemctl daemon-reload
`))

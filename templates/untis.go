package templates

import "text/template"

var ProjectUnitTemplate = template.Must(template.New("").Parse(`[Unit]
Description={{.Name}}

[Service]
Type=oneshot
ExecStart=/bin/true
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`))

var ServiceUnitTemplate = template.Must(template.New("").Parse(`[Unit]
Description={{.Name}} as part of {{.Project.Name}}
PartOf={{.Project.Slug}}.service
After={{.Project.Slug}}.service

[Service]
{{- with .Envs}}
{{- range .}}
Environment={{.}}
{{- end}}
{{- end}}
ExecStart={{.Binary}}
Restart={{.Restart}}
RestartSec={{.RestartSec}}

[Install]
WantedBy={{.Project.Slug}}.service
`))

var OneShotUnitTemplate = template.Must(template.New("").Parse(`[Unit]
Description={{.Name}}

[Service]
Type=oneshot
ExecStart={{.Binary}}

[Install]
WantedBy=multi-user.target
`))

var TimerUnitTemplate = template.Must(template.New("").Parse(`[Unit]
Description=Timer {{.Name}} as part of {{.Project.Name}}
PartOf={{.Project.Slug}}.service
After={{.Project.Slug}}.service
Requires={{.Launcher.Slug}}.service

[Timer]
Unit={{.Launcher.Slug}}.service
OnUnitInactiveSec={{.Interval}}

[Install]
WantedBy={{.Project.Slug}}.service
`))

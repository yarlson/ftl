package proxy

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/yarlson/ftl/pkg/config"
)

// GenerateNginxConfig generates an Nginx configuration based on the provided config.
func GenerateNginxConfig(cfg *config.Config) (string, error) {
	if cfg.Project.Domain == "" {
		cfg.Project.Domain = "localhost"
	}

	tmpl := template.Must(template.New("nginx").Parse(`
{{- range .Services}}
	upstream {{.Name}} {
		server {{.Name}}:{{.Port}};
	}
{{- end}}

	server {
		listen 443 ssl;
		http2 on;
		server_name {{.Project.Domain}};

		ssl_certificate /etc/nginx/certs/{{.Project.Domain}}.crt;
		ssl_certificate_key /etc/nginx/certs/{{.Project.Domain}}.key;
		ssl_protocols TLSv1.2 TLSv1.3;
		ssl_prefer_server_ciphers on;

        client_body_buffer_size 10M;
        client_max_body_size 10M;

        proxy_request_buffering off;

        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
{{- range .Services}}
	{{- $serviceName := .Name }}
	{{- range .Routes}}
		location {{.PathPrefix}} {
		{{- if .StripPrefix}}
			rewrite ^{{.PathPrefix}}(.*)$ /$1 break;
		{{- end}}
			resolver 127.0.0.11 valid=1s;
			set $service {{$serviceName}};
			proxy_pass http://$service;            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
		}
	{{- end}}
{{- end}}
	}
`))

	var buffer bytes.Buffer

	err := tmpl.Execute(&buffer, cfg)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(buffer.String(), "\t", "    "), nil
}

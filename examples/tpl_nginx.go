package main

import (
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/utils"
)

type Nginx struct {
	Version            string `urknall:"default=1.4.4"`
	HeadersMoreVersion string `urknall:"default=0.24"`
	SyslogPatchVersion string `urknall:"default=1.3.14"`
	Local              bool   // install to /usr/local/nginx
	Autostart          bool
}

func (pkg *Nginx) Render(r urknall.Package) {
	syslogPatchPath := "/tmp/nginx_syslog_patch"
	fileName := "syslog_{{ .SyslogPatchVersion }}.patch"
	r.AddCommands("base",
		InstallPackages("build-essential", "curl", "libpcre3", "libpcre3-dev", "libssl-dev", "libpcrecpp0", "zlib1g-dev", "libgd2-xpm-dev"),
		DownloadAndExtract(pkg.url(), "/opt/src"),
		Mkdir(syslogPatchPath, "root", 0755),
		Download("https://raw.github.com/yaoweibin/nginx_syslog_patch/master/config", syslogPatchPath+"/config", "root", 0644),
		Download("https://raw.github.com/yaoweibin/nginx_syslog_patch/master/"+fileName, syslogPatchPath+"/"+fileName, "root", 0644),
		And(
			"cd /opt/src/nginx-{{ .Version }}",
			"patch -p1 < "+syslogPatchPath+"/"+fileName,
		),
		Download("https://github.com/agentzh/headers-more-nginx-module/archive/v{{ .HeadersMoreVersion }}.tar.gz", "/opt/src/headers-more-nginx-module-{{ .HeadersMoreVersion }}.tar.gz", "root", 0644),
		And(
			"cd /opt/src",
			"tar xvfz headers-more-nginx-module-{{ .HeadersMoreVersion }}.tar.gz",
		),
		And(
			"cd /opt/src/nginx-{{ .Version }}",
			"./configure --with-http_ssl_module --with-http_gzip_static_module --with-http_stub_status_module --with-http_spdy_module --add-module=/tmp/nginx_syslog_patch --add-module=/opt/src/headers-more-nginx-module-{{ .HeadersMoreVersion }} --prefix={{ .InstallDir }}",
			"make",
			"make install",
		),
		WriteFile("/etc/init/nginx.conf", utils.MustRenderTemplate(upstartScript, pkg), "root", 0644),
	)
}

func (pkg *Nginx) InstallDir() string {
	if pkg.Local {
		return "/usr/local/nginx"
	}
	if pkg.Version == "" {
		panic("Version must be set")
	}
	return "/opt/nginx-" + pkg.Version
}

func (pkg *Nginx) BinPath() string {
	return pkg.InstallDir() + "/sbin/nginx"
}

func (pkg *Nginx) ReloadCommand() string {
	return utils.MustRenderTemplate("{{ . }} -t && {{ . }} -s reload", pkg.BinPath())
}

const upstartScript = `# nginx
 
description "nginx http daemon"
author "George Shammas <georgyo@gmail.com>"
 
{{ if .Autostart }}
start on (filesystem and net-device-up IFACE=lo)
stop on runlevel [!2345]
{{ end }}
 
env DAEMON={{ .InstallDir }}/sbin/nginx
env PID=/var/run/nginx.pid
 
respawn
respawn limit 10 5
#oom never
 
pre-start script
        $DAEMON -t
        if [ $? -ne 0 ]
                then exit $?
        fi
end script
 
exec $DAEMON
`

func (pkg *Nginx) url() string {
	return "http://nginx.org/download/" + pkg.fileName()
}

func (pkg *Nginx) fileName() string {
	return pkg.name() + ".tar.gz"
}

func (pkg *Nginx) name() string {
	return "nginx-" + pkg.Version
}

package server

import (
	"net/http"
	"opensavecloudserver/config"
	"opensavecloudserver/constant"
	"runtime"
)

type information struct {
	AllowRegister  bool   `json:"allow_register"`
	Version        string `json:"version"`
	ApiVersion     int    `json:"api_version"`
	GoVersion      string `json:"go_version"`
	OsName         string `json:"os_name"`
	OsArchitecture string `json:"os_architecture"`
}

func Information(w http.ResponseWriter, r *http.Request) {
	info := information{
		AllowRegister:  config.Features().AllowRegister,
		Version:        constant.Version,
		ApiVersion:     constant.ApiVersion,
		GoVersion:      runtime.Version(),
		OsName:         runtime.GOOS,
		OsArchitecture: runtime.GOARCH,
	}
	ok(info, w, r)
}

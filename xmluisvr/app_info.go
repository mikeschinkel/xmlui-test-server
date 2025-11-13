package xmluisvr

import (
	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
)

var appInfo = appinfo.New(appinfo.Args{
	AppName:    common.AppName,
	AppDescr:   common.AppDescr,
	AppSlug:    common.AppSlug,
	ConfigSlug: common.ConfigSlug,
	ConfigFile: common.ConfigFile,
	AppVer:     common.AppVer,
	InfoURL:    common.InfoURL,
	ExeName:    common.ExeName,
	LogFile:    common.LogFile,
	LogPath:    common.LogPath,
	ExtraInfo:  common.ExtraInfo,
})

func AppInfo() appinfo.AppInfo {
	return appInfo
}

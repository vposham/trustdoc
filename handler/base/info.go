package base

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// GitCommitId will be injected during go build in Dockerfile
	GitCommitId   string
	appInfoLoaded = false
	podId         string
	startTs       = time.Now().Format(time.RFC1123)
	appInfo       *AppInfo
)

// AppInfo is the response which contains app information
type AppInfo struct {
	GitCommitId string `json:"lastCommitId"`
	K8PodId     string `json:"kubePodId"`
	UpSince     string `json:"upSince"`
}

// GetInfo returns app info like app name, git commitId, etc.
func GetInfo(c *gin.Context) {
	c.JSON(http.StatusOK, getAppInfo())
}

func getAppInfo() AppInfo {
	if !appInfoLoaded {
		podId, _ = os.Hostname()
		appInfo = &AppInfo{
			GitCommitId: GitCommitId,
			K8PodId:     podId,
			UpSince:     startTs,
		}
		appInfoLoaded = true
	}
	return *appInfo
}

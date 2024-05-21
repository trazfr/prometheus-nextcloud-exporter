package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	nextCloudYes = `"yes"`
	nextCloudNo  = `"no"`
)

type NextCloudRoot struct {
	Ocs NextCloudOcs `json:"ocs"`
}

type NextCloudOcs struct {
	Data NextCloudData `json:"data"`
	Meta NextCloudMeta `json:"meta"`
}

type NextCloudData struct {
	ActiveUsers NextCloudActiveUsers `json:"activeUsers"`
	Nextcloud   NextcloudStruct      `json:"nextcloud"`
	Server      NextCloudServer      `json:"server"`
}

type NextCloudActiveUsers struct {
	Last1Day     int `json:"last24hours"`
	Last24Hour   int `json:"last1hour"`
	Last5Minutes int `json:"last5minutes"`
}

type NextCloudMeta struct {
	Message    string `json:"message"`
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
}

type NextcloudStruct struct {
	Shares  NextCloudShares  `json:"shares"`
	Storage NextCloudStorage `json:"storage"`
	System  NextCloudSystem  `json:"system"`
}

type NextCloudServer struct {
	Database  NextCloudDatabase `json:"database"`
	PHP       NextCloudPHP      `json:"php"`
	Webserver string            `json:"webserver"`
}

type NextCloudShares struct {
	NumFedSharesReceived    int `json:"num_fed_shares_received"`
	NumFedSharesSent        int `json:"num_fed_shares_sent"`
	NumShares               int `json:"num_shares"`
	NumSharesGroups         int `json:"num_shares_groups"`
	NumSharesLink           int `json:"num_shares_link"`
	NumSharesLinkNoPassword int `json:"num_shares_link_no_password"`
	NumSharesMail           int `json:"num_shares_mail"`
	NumSharesRoom           int `json:"num_shares_room"`
	NumSharesUser           int `json:"num_shares_user"`
}

type NextCloudStorage struct {
	NumUsers         int `json:"num_users"`
	NumFiles         int `json:"num_files"`
	NumStorages      int `json:"num_storages"`
	NumStoragesLocal int `json:"num_storages_local"`
	NumStoragesHome  int `json:"num_storages_home"`
	NumStoragesOther int `json:"num_storages_other"`
}

type NextCloudSystem struct {
	Apps                *NextCloudApps   `json:"apps"`
	CPULoad             []float64        `json:"cpuload"`
	Debug               NextCloudYesNo   `json:"debug"`
	EnableAvatars       NextCloudYesNo   `json:"enable_avatars"`
	EnablePreviews      NextCloudYesNo   `json:"enable_previews"`
	FilelockingEnabled  NextCloudYesNo   `json:"filelocking.enabled"`
	FreeSpace           int64            `json:"freespace"`
	MemFree             int64            `json:"mem_free"`
	MemTotal            int64            `json:"mem_total"`
	MemcacheDistributed string           `json:"memcache.distributed"`
	MemcacheLocal       string           `json:"memcache.local"`
	MemcacheLocking     string           `json:"memcache.locking"`
	SwapFree            int64            `json:"swap_free"`
	SwapTotal           int64            `json:"swap_total"`
	Theme               string           `json:"theme"`
	Update              *NextcloudUpdate `json:"update"`
	Version             string           `json:"version"`
}

type NextCloudDatabase struct {
	Size    NextCloudIntOrString `json:"size"`
	Type    string               `json:"type"`
	Version string               `json:"version"`
}

type NextCloudPHP struct {
	MaxExecutionTime  int    `json:"max_execution_time"`
	MemoryLimit       int64  `json:"memory_limit"`
	UploadMaxFileSize int64  `json:"upload_max_filesize"`
	Version           string `json:"version"`
}

type NextCloudApps struct {
	//AppUpdates          map[string]string `json:"app_updates"`
	NumInstalled        int `json:"num_installed"`
	NumUpdatesAvailable int `json:"num_updates_available"`
}

type NextcloudUpdate struct {
	Available      bool  `json:"available"`
	LastUpdateDate int64 `json:"lastupdatedat"`
}

type NextCloudYesNo bool

func (n *NextCloudYesNo) UnmarshalJSON(data []byte) error {
	var err error
	switch string(data) {
	case nextCloudYes:
		*n = true
	case nextCloudNo:
		*n = false
	default:
		err = fmt.Errorf("cannot unmarshal: %v", data)
	}
	return err
}

type NextCloudIntOrString int64

func (n *NextCloudIntOrString) UnmarshalJSON(data []byte) error {
	if data[0] != '"' {
		return json.Unmarshal(data, (*int64)(n))
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*n = NextCloudIntOrString(i)
	return nil
}

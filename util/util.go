package util

import (
	"github.com/kenbragn/gosync-reborn/manifests"
	"net/url"
)

func MapFileListAsync(manifest []manifests.FileManifest, mapFileManifestChan chan map[string]manifests.FileManifest) {
	mp := make(map[string]manifests.FileManifest)
	for _, fileManifest := range manifest {
		mp[fileManifest.Path] = fileManifest
	}
	mapFileManifestChan <- mp
}

func MapFileListSync(manifest []manifests.FileManifest) map[string]manifests.FileManifest {
	mapFileManifest := make(map[string]manifests.FileManifest)
	for _, fileManifest := range manifest {
		mapFileManifest[fileManifest.Path] = fileManifest
	}
	return mapFileManifest
}

func IsURL(check string) bool {
	urlCheck, err := url.ParseRequestURI(check)
	if urlCheck.Host == "" && urlCheck.Scheme == "" {
		return false
	}
	if err != nil {
		return false
	} else {
		return true
	}
}

package manifests

import "time"

// ObsoleteFileManifest Obsolete file manifests element
type ObsoleteFileManifest struct {
	Path string `json:"path"`
}

// ZSyncFile manifests element
type ZSyncFile struct {
	UUID            string    `json:"uuid"`
	ZsyncVersion    string    `json:"zsyncVersion"`
	Filename        string    `json:"filename"`
	ModifiedTime    time.Time `json:"modifiedTime"`
	Blocksize       int       `json:"blockSize"`
	Length          int64     `json:"length"`
	SeqMatches      int       `json:"seqMatches"`
	WeaksumLength   int       `json:"weaksumLength"`
	StrongsumLength int       `json:"strongsumLength"`
	URL             string    `json:"url"`
	FileChecksum    string    `json:"fileChecksum"`
	Checksum        string    `json:"checksum"`
}

// FileManifest File manifests element
type FileManifest struct {
	Path     string          `json:"path"`
	FileSize uint64          `json:"filesize"`
	Blocks   []BlockManifest `json:"blocks"`
	Checksum string          `json:"checksum"`

	//V2
	ZSyncFile *ZSyncFile `json:"zsyncFile,omitempty"`
	UUID      *string    `json:"uuid,omitempty"`
}

// BlockManifest json for blocks elements
type BlockManifest struct {
	UUID      string `json:"uuid"`
	Checksum  string `json:"checksum"`
	Offset    uint64 `json:"offset"`
	BlockSize int64  `json:"blockSize"`
}

// DefaultLaunchProfile Manifest for app-specific configuration
type DefaultLaunchProfile struct {
	DefaultEntryPoint  string `json:"defaultEntryPoint"`
	DefaultClientID    string `json:"defaultClientId"`
	DefaultRedirectURI string `json:"defaultRedirectURI"`
}

// BuildManifest Patch manifests data
type BuildManifest struct {
	AppID                string                 `json:"appId,omitempty"`
	Version              string                 `json:"displayVersion,omitempty"`
	BuildID              string                 `json:"buildId"`
	PlatformID           string                 `json:"platformId"`
	BaseURLs             []string               `json:"baseUrls"`
	Files                []FileManifest         `json:"files"`
	DefaultLaunchProfile DefaultLaunchProfile   `json:"defaultLaunchProfile"`
	ObsoleteFiles        []ObsoleteFileManifest `json:"obsoleteFiles"`
}

type UploadSummary struct {
	PresignedUrl string `json:"presignedUrl"`
	Uuid         string `json:"uuid"`
}

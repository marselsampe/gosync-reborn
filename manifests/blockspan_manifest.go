package manifests

import "github.com/AccelByte/gosync-reborn/patcher"

type PatchingBlockSpan struct {
	MissingSpans []patcher.MissingBlockSpan `json:"missingBlockSpans"`
	FoundSpans   []patcher.FoundBlockSpan   `json:"foundBlockSpans"`
	TotalBlocks  uint                       `json:"totalBlocks"`
	FullDownload bool                       `json:"fullDownload"`
	NoDownload   bool                       `json:"noDownload"`
}

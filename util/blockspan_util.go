package util

import (
	"github.com/kenbragn/gosync-reborn/manifests"
	"github.com/kenbragn/gosync-reborn/patcher"
)

type Block struct {
	RemoteSource      bool
	StartBlock        uint
	EndBlock          uint
	MatchOffset       int64
	Offset            int64
	BlockSize         int64
	TotalBlockToFetch int64
}

func FormBlocksOrder(blockSpan manifests.PatchingBlockSpan) []Block {
	if blockSpan.NoDownload {
		// TODO copy from old file
		// return ?
	} else if blockSpan.FullDownload {
		// TODO upload all
		// return ?
	}

	maxBlock := blockSpan.TotalBlocks
	currentBlock := uint(0)
	var blocks []Block

	for currentBlock <= maxBlock {
		var b Block
		if withinFirstBlockOfLocalBlocks(currentBlock, blockSpan.FoundSpans) {
			firstMatched := blockSpan.FoundSpans[0]
			blockSizeToRead := int64(firstMatched.EndBlock-firstMatched.StartBlock+1) * firstMatched.BlockSize

			b.RemoteSource = false
			b.TotalBlockToFetch = blockSizeToRead
			b.BlockSize = firstMatched.BlockSize
			b.StartBlock = firstMatched.StartBlock
			b.EndBlock = firstMatched.EndBlock
			b.MatchOffset = firstMatched.MatchOffset
			b.Offset = int64(firstMatched.StartBlock) * firstMatched.BlockSize

			currentBlock = firstMatched.EndBlock + 1
			blockSpan.FoundSpans = blockSpan.FoundSpans[1:]
		} else if withinFirstBlockOfRemoteBlocks(currentBlock, blockSpan.MissingSpans) {
			firstMissing := blockSpan.MissingSpans[0]
			blockSizeToRead := int64(firstMissing.EndBlock-firstMissing.StartBlock+1) * firstMissing.BlockSize
			startOffset := int64(firstMissing.StartBlock) * firstMissing.BlockSize
			endOffset := int64(startOffset) + blockSizeToRead - 1
			responseLen := endOffset - startOffset + 1
			completed := calculateNumberOfCompletedBlocks(responseLen, firstMissing.BlockSize)

			b.RemoteSource = true
			b.TotalBlockToFetch = blockSizeToRead
			b.BlockSize = firstMissing.BlockSize
			b.StartBlock = firstMissing.StartBlock
			b.EndBlock = firstMissing.EndBlock
			b.Offset = int64(firstMissing.StartBlock) * firstMissing.BlockSize

			currentBlock += uint(completed)
			blockSpan.MissingSpans = blockSpan.MissingSpans[1:]
		}
		blocks = append(blocks, b)
	}
	return blocks

}

func withinFirstBlockOfLocalBlocks(currentBlock uint, localBlocks []patcher.FoundBlockSpan) bool {
	return len(localBlocks) > 0 && localBlocks[0].StartBlock <= currentBlock && localBlocks[0].EndBlock >= currentBlock
}

func withinFirstBlockOfRemoteBlocks(currentBlock uint, remoteBlocks []patcher.MissingBlockSpan) bool {
	return len(remoteBlocks) > 0 && remoteBlocks[0].StartBlock <= currentBlock && remoteBlocks[0].EndBlock >= currentBlock
}

func calculateNumberOfCompletedBlocks(resultLength int64, blockSize int64) int64 {

	completedBlockCount := resultLength / blockSize

	// round up in the case of a partial block (last block may not be full sized)
	if resultLength%blockSize != 0 {
		completedBlockCount += 1
	}

	return completedBlockCount
}


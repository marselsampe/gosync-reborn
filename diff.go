package gosync

import (
	"errors"
	"fmt"
	"github.com/AccelByte/gosync-reborn/manifests"
	"github.com/AccelByte/gosync-reborn/patcher"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
)

func CalculateDiff(localFilePath string, summaryFilePath string, baseLocalPath string) (patchingBlockSpan *manifests.PatchingBlockSpan, errCalculateDiff error) {
	remoteReferenceFileURL := "http://localhost"
	outputFilePath := path.Join( baseLocalPath, "inexistentoutputfilepath")

	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintln(os.Stderr, p)
			patchingBlockSpan = nil
			errCalculateDiff = errors.New("Recover from panic")
		}
	}()

	span, errDiff := doCalculateDiff(localFilePath, summaryFilePath, remoteReferenceFileURL, outputFilePath)
	if errDiff != nil {
		return nil, errDiff
	}

	return span, nil
}

func GenerateFullMissingBytesDiff(summaryFilePath string) (patchingBlockSpan *manifests.PatchingBlockSpan, errFullMissingBytes error) {
	basicSummary, errCreateBasicSummary := createBasicSummary(summaryFilePath)
	if errCreateBasicSummary != nil {
		return nil, errFullMissingBytes
	}
	fileSize := basicSummary.FileSize
	blockSize := basicSummary.BlockSize
	division := fileSize/int64(blockSize)
	totalBlocks := math.Ceil(float64(division))

	var missingSpans []patcher.MissingBlockSpan
	singleMissingSpan := patcher.MissingBlockSpan{StartBlock:0, EndBlock:uint(totalBlocks), BlockSize:int64(blockSize)}
	missingSpans = append(missingSpans, singleMissingSpan)

	blockSpan := &manifests.PatchingBlockSpan{MissingSpans:missingSpans, TotalBlocks:uint(totalBlocks), FullDownload:true, NoDownload:false}

	return blockSpan, nil

}

func doCalculateDiff(localFilename string, summaryFile string, remoteReferenceURL string, outputFilePath string) (*manifests.PatchingBlockSpan, error) {

	outFilename := localFilename
	if outputFilePath != "" {
		outFilename = outputFilePath
	}

	basicSummary, errCreateBasicSummary := createBasicSummary(summaryFile)
	if errCreateBasicSummary != nil {
		return nil, errCreateBasicSummary
	}

	rsync, err := MakeRSync(
		localFilename,
		remoteReferenceURL,
		outFilename,
		basicSummary,
	)

	if err != nil {
		return nil, err
	}
	defer rsync.Close()

	patchingBlockSpan, errCalc := rsync.CalculateDiffV2()
	if errCalc != nil {
		return nil, errCalc
	}

	return patchingBlockSpan, nil
}

func createBasicSummary(summaryFile string) (summary *BasicSummary, errCreateBasicSummary error) {
	if isURL(summaryFile) {
		tempZsyncFile, errDownloadZsyncFile := downloadFile(summaryFile)
		if errDownloadZsyncFile != nil {
			fmt.Printf("Error downloading file: %v", errDownloadZsyncFile)
			return nil, errDownloadZsyncFile
		}
		summaryFile = tempZsyncFile
	}

	indexReader, e := os.Open(summaryFile)
	if e != nil {
		fmt.Printf("Error opening file: %s", e)
		return nil, e
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing zsync temp file: %v", e)
		}
	}(indexReader)

	_, _, _, filesize, blocksize, errReadHeadersAndCheck := readHeadersAndCheck(
		indexReader,
		magicString,
		majorVersion,
	)

	if errReadHeadersAndCheck != nil {
		return nil, errReadHeadersAndCheck
	}

	index, checksumLookup, blockCount, errReadIndex := readIndex(
		indexReader,
		uint(blocksize),
	)

	if errReadIndex != nil {
		return nil, errReadIndex
	}

	basicSummary := &BasicSummary{
		ChecksumIndex:  index,
		ChecksumLookup: checksumLookup,
		BlockCount:     blockCount,
		BlockSize:      uint(blocksize),
		FileSize:       filesize,
	}
	return basicSummary, nil
}

func downloadFile(url string) (tempFilePath string, err error) {

	// Create the file
	out, err := ioutil.TempFile("", "gosync")
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}

func isURL(check string) bool {
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

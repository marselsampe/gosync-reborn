package gosync

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/AccelByte/gosync-reborn/blocksources"
	"github.com/AccelByte/gosync-reborn/chunks"
	"github.com/AccelByte/gosync-reborn/comparer"
	"github.com/AccelByte/gosync-reborn/filechecksum"
	"github.com/AccelByte/gosync-reborn/index"
	"github.com/AccelByte/gosync-reborn/manifests"
	"github.com/AccelByte/gosync-reborn/patcher"
	"github.com/AccelByte/gosync-reborn/patcher/sequential"
	"github.com/AccelByte/gosync-reborn/util"
)

const (
	megabyte = 1000000
)

// ReadSeekerAt is the combinaton of ReadSeeker and ReaderAt interfaces
type ReadSeekerAt interface {
	io.ReadSeeker
	io.ReaderAt
}

/*
RSync is an object designed to make the standard use-case for gosync as
easy as possible.

To this end, it hides away many low level choices by default, and makes some
assumptions.
*/
type RSync struct {
	Input  ReadSeekerAt
	Source patcher.BlockSource
	Output io.Writer

	Summary FileSummary

	OnClose []closer

	Concurrency int // Concurrency is the concurrency level used by diffing, patching and downloading
}

type closer interface {
	Close() error
}

// FileSummary combines many of the interfaces that are needed
// It is expected that you might implement it by embedding existing structs
type FileSummary interface {
	GetBlockSize() uint
	GetBlockCount() uint
	GetFileSize() int64
	FindWeakChecksum2(bytes []byte) interface{}
	FindStrongChecksum2(bytes []byte, weak interface{}) []chunks.ChunkChecksum
	GetStrongChecksumForBlock(blockID int) []byte
}

// BasicSummary implements a version of the FileSummary interface
type BasicSummary struct {
	BlockSize  uint
	BlockCount uint
	FileSize   int64
	*index.ChecksumIndex
	filechecksum.ChecksumLookup
}

// GetBlockSize gets the size of each block
func (fs *BasicSummary) GetBlockSize() uint {
	return fs.BlockSize
}

// GetBlockCount gets the number of blocks
func (fs *BasicSummary) GetBlockCount() uint {
	return fs.BlockCount
}

// GetFileSize gets the file size of the file
func (fs *BasicSummary) GetFileSize() int64 {
	return fs.FileSize
}

// MakeRSync creates an RSync object using string paths,
// inferring most of the configuration
func MakeRSync(
	InputFile,
	Source,
	OutFile string,
	Summary FileSummary,
	Concurrency int,
) (r *RSync, err error) {
	useTempFile := false
	if useTempFile, err = IsSameFile(InputFile, OutFile); err != nil {
		return nil, err
	}

	inputFile, err := os.Open(InputFile)

	if err != nil {
		return
	}

	var out io.WriteCloser
	var outFilename = OutFile
	var copier closer

	if useTempFile {
		out, outFilename, err = getTempFile()

		if err != nil {
			return
		}

		copier = &fileCopyCloser{
			from: outFilename,
			to:   OutFile,
		}
	} else {
		out, err = getOutFile(OutFile)

		if err != nil {
			return
		}

		copier = nullCloser{}
	}

	// blocksource
	var source *blocksources.BlockSourceBase

	resolver := blocksources.MakeFileSizedBlockResolver(
		uint64(Summary.GetBlockSize()),
		Summary.GetFileSize(),
	)

	source = blocksources.NewHttpBlockSource(
		Source,
		Concurrency,
		resolver,
		&filechecksum.HashVerifier{
			Hash:                md5.New(),
			BlockSize:           Summary.GetBlockSize(),
			BlockChecksumGetter: Summary,
		},
	)

	r = &RSync{
		Input:   inputFile,
		Output:  out,
		Source:  source,
		Summary: Summary,
		OnClose: []closer{
			&fileCloser{inputFile, InputFile},
			&fileCloser{out, outFilename},
			copier,
		},
		Concurrency: Concurrency,
	}

	return
}

// Patch the files
func (rsync *RSync) Patch() (err error) {
	numMatchers := int64(rsync.Concurrency)
	blockSize := rsync.Summary.GetBlockSize()
	sectionSize := rsync.Summary.GetFileSize() / numMatchers
	sectionSize += int64(blockSize) - (sectionSize % int64(blockSize))

	merger := &comparer.MatchMerger{}

	for i := int64(0); i < numMatchers; i++ {
		compare := &comparer.Comparer{}
		offset := sectionSize * i

		sectionReader := bufio.NewReaderSize(
			io.NewSectionReader(rsync.Input, offset, sectionSize+int64(blockSize)),
			megabyte, // 1 MB buffer
		)

		// Bakes in the assumption about how to generate checksums (extract)
		sectionGenerator := filechecksum.NewFileChecksumGenerator(
			uint(blockSize),
		)

		matchStream := compare.StartFindMatchingBlocks(
			sectionReader, offset, sectionGenerator, rsync.Summary,
		)

		merger.StartMergeResultStream(matchStream, int64(blockSize))
	}

	mergedBlocks := merger.GetMergedBlocks()
	missing := mergedBlocks.GetMissingBlocks(rsync.Summary.GetBlockCount() - 1)

	return sequential.SequentialPatcher(
		rsync.Input,
		rsync.Source,
		toPatcherMissingSpan(missing, int64(blockSize)),
		toPatcherFoundSpan(mergedBlocks, int64(blockSize)),
		20*megabyte,
		rsync.Output,
	)
}


func (rsync *RSync) CalculateDiffAndMarshall() (diff string, err error) {

	retVal, errCalc := rsync.CalculateDiffV2()
	if errCalc != nil {
		return "", errCalc
	}

	jsonVal, errJson := json.Marshal(retVal)

	if errJson != nil {
		return "", errJson
	}

	jsonStr := string(jsonVal)

	return jsonStr, nil
}

func (rsync *RSync) CalculateDiffV2() (diff *manifests.PatchingBlockSpan, err error) {
	numMatchers := int64(rsync.Concurrency)
	blockSize := rsync.Summary.GetBlockSize()
	sectionSize := rsync.Summary.GetFileSize() / numMatchers
	sectionSize += int64(blockSize) - (sectionSize % int64(blockSize))

	merger := &comparer.MatchMerger{}

	for i := int64(0); i < numMatchers; i++ {
		compare := &comparer.Comparer{}
		offset := sectionSize * i

		sectionReader := bufio.NewReaderSize(
			io.NewSectionReader(rsync.Input, offset, sectionSize+int64(blockSize)),
			megabyte, // 1 MB buffer
		)

		// Bakes in the assumption about how to generate checksums (extract)
		sectionGenerator := filechecksum.NewFileChecksumGenerator(
			uint(blockSize),
		)

		matchStream := compare.StartFindMatchingBlocks(
			sectionReader, offset, sectionGenerator, rsync.Summary,
		)

		merger.StartMergeResultStream(matchStream, int64(blockSize))
	}

	mergedBlocks := merger.GetMergedBlocks()
	missing := mergedBlocks.GetMissingBlocks(rsync.Summary.GetBlockCount() - 1)

	missingSpan := toPatcherMissingSpan(missing, int64(blockSize))
	foundSpan := toPatcherFoundSpan(mergedBlocks, int64(blockSize))

	var totalBlocks uint
	if len(missingSpan) > 0 {
		totalBlocks = missingSpan[len(missingSpan)-1].EndBlock
	}

	if len(foundSpan) > 0 && foundSpan[len(foundSpan)-1].EndBlock > totalBlocks {
		totalBlocks = foundSpan[len(foundSpan)-1].EndBlock
	}

	var splitMissingSpans []patcher.MissingBlockSpan
	maxSpanLength := 10000

	for i := 0; i < len(missingSpan); i++ {
		missing := missingSpan[i]
		var splitSingleMissingSpan []patcher.MissingBlockSpan
		smallSpans := util.SplitSpan(missing, splitSingleMissingSpan, uint(maxSpanLength))
		splitMissingSpans = append(splitMissingSpans, smallSpans...)
	}

	if splitMissingSpans == nil {
		splitMissingSpans = make([]patcher.MissingBlockSpan, 0)
	}

	return &manifests.PatchingBlockSpan{MissingSpans: splitMissingSpans, FoundSpans: foundSpan, TotalBlocks: totalBlocks}, nil
}

func getOutFile(filename string) (f io.WriteCloser, err error) {
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	}

	return os.OpenFile(filename, os.O_WRONLY, 0)
}

func getTempFile() (f io.WriteCloser, filename string, err error) {
	ft, err := ioutil.TempFile(".", "tmp_")
	filename = ft.Name()
	f = ft
	return
}

// IsSameFile checks if two file paths are the same file
func IsSameFile(path1, path2 string) (same bool, err error) {

	fi1, err := os.Stat(path1)

	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return
	}

	fi2, err := os.Stat(path2)

	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return
	}

	return os.SameFile(fi1, fi2), nil
}

// Close - close open files, copy to the final location from
// a temporary one if neede
func (rsync *RSync) Close() error {
	for _, f := range rsync.OnClose {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

type fileCloser struct {
	f    io.Closer
	path string
}

// Close - add file path information to closing a file
func (f *fileCloser) Close() error {
	err := f.f.Close()
	if err != nil {
		return fmt.Errorf(
			"Could not close file %v: %v",
			f.path,
			err,
		)
	}
	return nil
}

type nullCloser struct{}

func (n nullCloser) Close() error {
	return nil
}

type fileCopyCloser struct {
	from string
	to   string
}

func (f *fileCopyCloser) Close() (err error) {
	from, err := os.OpenFile(f.from, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer func() {
		e := from.Close()
		if err != nil {
			err = e
		}
	}()

	to, err := os.OpenFile(f.to, os.O_TRUNC|os.O_WRONLY, 0)

	if err != nil {
		return err
	}

	defer func() {
		e := to.Close()
		if err != nil {
			err = e
		}
	}()

	bufferedReader := bufio.NewReaderSize(from, megabyte)
	_, err = io.Copy(to, bufferedReader)
	return
}
package util

import (
	"github.com/kenbragn/gosync-reborn/patcher"
	"github.com/magiconair/properties/assert"
	"testing"
)

const (
	BLOCKSIZE        = 4
	MAX_ALLOWED_SPAN_LENGTH = 12
)

func TestSplitSpan_normal(t *testing.T) {
	originalSpan := patcher.MissingBlockSpan{StartBlock: 0, EndBlock: BLOCKSIZE * 24}
	expectedSpanLength := 8
	var newArrSpan []patcher.MissingBlockSpan
	span := SplitSpan(originalSpan, newArrSpan, MAX_ALLOWED_SPAN_LENGTH)
	assert.Equal(t, len(span), expectedSpanLength, "Number of blocks")
	for i := 0; i < len(span); i++ {
		testSpan := span[i]
		assert.Equal(t, testSpan.StartBlock, uint(i*MAX_ALLOWED_SPAN_LENGTH), "Start block")
		if i == len(span) - 1 {	// eof, might be less on the endBlock
			assert.Equal(t, testSpan.EndBlock, originalSpan.EndBlock, "End block")
		} else {
			assert.Equal(t, testSpan.EndBlock, testSpan.StartBlock + uint(MAX_ALLOWED_SPAN_LENGTH ) - 1, "End block")
		}
	}
}

func TestSplitSpan_edge1(t *testing.T) {
	originalSpan := patcher.MissingBlockSpan{StartBlock: 0, EndBlock: BLOCKSIZE}
	expectedSpanLength := 1
	var newArrSpan []patcher.MissingBlockSpan
	span := SplitSpan(originalSpan, newArrSpan, MAX_ALLOWED_SPAN_LENGTH)
	assert.Equal(t, len(span), expectedSpanLength, "Number of blocks")
	for i := 0; i < len(span); i++ {
		testSpan := span[i]
		assert.Equal(t, testSpan.StartBlock, uint(i*MAX_ALLOWED_SPAN_LENGTH), "Start block")
		if i == len(span) - 1 {	// eof, might be less on the endBlock
			assert.Equal(t, testSpan.EndBlock, originalSpan.EndBlock, "End block")
		} else {
			assert.Equal(t, testSpan.EndBlock, testSpan.StartBlock + uint(MAX_ALLOWED_SPAN_LENGTH ) - 1, "End block")
		}
	}
}

func TestSplitSpan_edge2(t *testing.T) {
	originalSpan := patcher.MissingBlockSpan{StartBlock: 0, EndBlock: BLOCKSIZE * 10}
	expectedSpanLength := 4
	var newArrSpan []patcher.MissingBlockSpan
	span := SplitSpan(originalSpan, newArrSpan, MAX_ALLOWED_SPAN_LENGTH)
	assert.Equal(t, len(span), expectedSpanLength, "Number of blocks")
	for i := 0; i < len(span); i++ {
		testSpan := span[i]
		assert.Equal(t, testSpan.StartBlock, uint(i*MAX_ALLOWED_SPAN_LENGTH), "Start block")
		if i == len(span) - 1 {	// eof, might be less on the endBlock
			assert.Equal(t, testSpan.EndBlock, originalSpan.EndBlock, "End block")
		} else {
			assert.Equal(t, testSpan.EndBlock, testSpan.StartBlock + uint(MAX_ALLOWED_SPAN_LENGTH ) - 1, "End block")
		}
	}
}
package util

import "github.com/AccelByte/gosync-reborn/patcher"

func SplitSpan(span patcher.MissingBlockSpan, newArrSpan []patcher.MissingBlockSpan, maxAllowedSpan uint) []patcher.MissingBlockSpan {
	if span.EndBlock - span.StartBlock <= maxAllowedSpan {
		newArrSpan = append(newArrSpan, span)
		return newArrSpan
	} else {
		endSpan := span.StartBlock + maxAllowedSpan - 1
		smallerSpan := patcher.MissingBlockSpan{StartBlock:span.StartBlock, EndBlock:endSpan, BlockSize:span.BlockSize}
		nextSpan := patcher.MissingBlockSpan{StartBlock:endSpan + 1, EndBlock:span.EndBlock, BlockSize:span.BlockSize}
		newArrSpan = append(newArrSpan, smallerSpan)
		return SplitSpan(nextSpan, newArrSpan, maxAllowedSpan)
	}

}
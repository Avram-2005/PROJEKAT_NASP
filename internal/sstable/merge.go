package sstable

import (
	"container/heap"
	"fmt"

	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
)

func findMinMaxKeys(ssts []*SSTable) (string, string) {
	var minKey, maxKey string
	for _, sst := range ssts {
		if minKey == "" || sst.summary.firstKey < minKey {
			minKey = sst.summary.firstKey
		}
		if maxKey == "" || sst.summary.lastKey > maxKey {
			maxKey = sst.summary.lastKey
		}
	}
	return minKey, maxKey
}

type IterHeap []*SSTableIterator

func (h IterHeap) Len() int { return len(h) }

func (h IterHeap) Less(i, j int) bool {
	if h[i].Rec.Key != h[j].Rec.Key {
		return h[i].Rec.Key < h[j].Rec.Key
	}
	return h[i].Rec.Timestamp.After(h[j].Rec.Timestamp)
}
func (h IterHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *IterHeap) Push(x any) {
	*h = append(*h, x.(*SSTableIterator))
}

func (h *IterHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func newIterHeap(ssts []*SSTable, sstm *SSTableManager) (*IterHeap, error) {
	h := &IterHeap{}
	heap.Init(h)
	for _, sst := range ssts {
		iter, err := sstm.NewSSTableIterator(sst)
		if err != nil {
			return nil, fmt.Errorf("failed to create iterator for SSTable: %v", err)
		}
		heap.Push(h, iter)
	}
	return h, nil
}

func (h *IterHeap) Close() error {
	for _, iter := range *h {
		if err := iter.Close(); err != nil {
			return fmt.Errorf("failed to close iterator: %v", err)
		}
	}
	return nil
}

func reconstructFilter(sstm *SSTableManager, sst *SSTable, numRecs int) error {
	bf, err := BloomFilter.NewBloomFilter(uint(numRecs), BLOOM_FILTER_RATE)
	if err != nil {
		return fmt.Errorf("failed to create Bloom filter: %v", err)
	}

	iter, err := sstm.NewSSTableIterator(sst)
	if err != nil {
		return fmt.Errorf("failed to create iterator for new SSTable: %v", err)
	}

	for iter.Rec != nil {
		bf.Set([]byte(iter.Rec.Key))
		if _, err := iter.Next(); err != nil {
			return fmt.Errorf("failed to advance iterator: %v", err)
		}
	}
	sst.filter = bf
	return nil
}

func (sstm *SSTableManager) multipleFilesMerge(ssts []*SSTable, level int, tableNum int) (*SSTable, error) {
	// Cannot calculate number of records in advance, so we set it to 0 for now
	state, err := sstm.multipleFilesFlushInit(level, tableNum, 0)
	if err != nil {
		return nil, err
	}

	minKey, maxKey := findMinMaxKeys(ssts)
	state.summary.SetFirstAndLast(minKey, maxKey)
	writeSummaryHeader(state.summaryWriter, minKey, maxKey)

	h, err := newIterHeap(ssts, sstm)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator heap: %v", err)
	}

	numRecs := 0
	for h.Len() > 0 {
		minIter := heap.Pop(h).(*SSTableIterator)
		currentRec := minIter.Rec

		shouldWriteSummary := numRecs%sstm.config.SummaryInterval == 0
		sstm.multipleFilesFlushRecord(*currentRec, state, shouldWriteSummary)
		numRecs++

		if hasNext, err := minIter.Next(); err != nil {
			return nil, fmt.Errorf("failed to advance iterator: %v", err)
		} else if hasNext {
			heap.Push(h, minIter)
		}
	}
	sst, err := sstm.multipleFilesFlushFinalize(level, state, tableNum)

	reconstructFilter(sstm, sst, numRecs)

	filterData := sst.filter.Dump()
	state.filterWriter.Write(filterData)
	state.filterWriter.Finalize()

	for _, iter := range *h {
		if err := iter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close iterator: %v", err)
		}
	}

	return sst, err
}

func (sstm *SSTableManager) oneFileMerge(ssts []*SSTable, level int, tableNum int) (*SSTable, error) {
	// Cannot calculate number of records in advance, so we set it to 0 for now
	state, err := sstm.oneFileFlushInit(level, tableNum, 0)
	if err != nil {
		return nil, err
	}

	h, err := newIterHeap(ssts, sstm)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator heap: %v", err)
	}

	numRecs := 0
	for h.Len() > 0 {
		minIter := heap.Pop(h).(*SSTableIterator)
		currentRec := minIter.Rec

		sstm.oneFileFlushRecord(level, *currentRec, state)
		numRecs++

		if hasNext, err := minIter.Next(); err != nil {
			return nil, fmt.Errorf("failed to advance iterator: %v", err)
		} else if hasNext {
			heap.Push(h, minIter)
		}
	}
	sst, err := sstm.oneFileFlushFinalize(level, state, tableNum)

	reconstructFilter(sstm, sst, numRecs)

	filterData := sst.filter.Dump()
	state.writer.Seek(int(sst.footer.FilterStart))
	state.writer.Write(filterData)

	sst.footer.FooterStart = state.writer.CurrOffset()
	sst.footer.Write(state.writer)
	state.writer.Finalize()

	for _, iter := range *h {
		if err := iter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close iterator: %v", err)
		}
	}

	return sst, err
}

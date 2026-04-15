package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	bitsbloom "github.com/bits-and-blooms/bloom/v3"
	qbloom "github.com/nireo/qbloom"
)

type samplePoint struct {
	itemsPerBit float64
	avgFP       float64
	minFP       float64
	maxFP       float64
}

func main() {
	var (
		numBits        = flag.Int("num-bits", 1<<12, "fixed bloom filter size in bits")
		probeBudget    = flag.Int("probe-budget", 2_000_000, "adaptive false-positive probe budget per point")
		trials         = flag.Int("trials", 16, "number of independent trials")
		granularity    = flag.Float64("granularity", 128, "controls how many x-axis points are measured")
		maxItemsPerBit = flag.Float64("max-items-per-bit", 0.125, "stop after reaching this items-per-bit load")
		outDir         = flag.String("out-dir", "Acc", "directory for output CSV files")
	)
	flag.Parse()

	if *numBits <= 0 {
		fmt.Fprintln(os.Stderr, "num-bits must be >= 1")
		os.Exit(1)
	}
	if *numBits%64 != 0 {
		fmt.Fprintln(os.Stderr, "num-bits must be a multiple of 64")
		os.Exit(1)
	}
	if *trials < 1 {
		fmt.Fprintln(os.Stderr, "trials must be >= 1")
		os.Exit(1)
	}
	if *probeBudget < 1 {
		fmt.Fprintln(os.Stderr, "probe-budget must be >= 1")
		os.Exit(1)
	}
	if *granularity <= 0 {
		fmt.Fprintln(os.Stderr, "granularity must be > 0")
		os.Exit(1)
	}
	if *maxItemsPerBit <= 0 {
		fmt.Fprintln(os.Stderr, "max-items-per-bit must be > 0")
		os.Exit(1)
	}

	trialRows := make([][]trialPoint, *trials)
	var wg sync.WaitGroup
	for trial := 0; trial < *trials; trial++ {
		wg.Add(1)
		go func(trial int) {
			defer wg.Done()
			trialRows[trial] = runTrial(trial, *trials, *numBits, *probeBudget, *granularity, *maxItemsPerBit)
		}(trial)
	}
	wg.Wait()

	qbloomRows, bitsRows := aggregateTrials(trialRows)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output directory: %v\n", err)
		os.Exit(1)
	}

	if err := writeSeries(filepath.Join(*outDir, "qbloom.csv"), qbloomRows); err != nil {
		fmt.Fprintf(os.Stderr, "write qbloom series: %v\n", err)
		os.Exit(1)
	}
	if err := writeSeries(filepath.Join(*outDir, "bits-and-blooms.csv"), bitsRows); err != nil {
		fmt.Fprintf(os.Stderr, "write bits-and-blooms series: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d points to %s\n", len(qbloomRows), *outDir)
}

type trialPoint struct {
	itemsPerBit float64
	qbloomFP    float64
	bitsFP      float64
}

type ticks struct {
	cur      int
	step     float64
	stepSize float64
}

func (t *ticks) next() int {
	t.cur += 1 << int(t.step)
	t.step += t.stepSize
	return t.cur
}

func runTrial(trial, numTrials, numBits, probeBudget int, granularity, maxItemsPerBit float64) []trialPoint {
	const halfUint64 = ^uint64(0) / 2
	memberBase := uint64(trial) * (halfUint64 / uint64(numTrials))
	nonMemberBase := memberBase + halfUint64
	qbNonMemberNext := nonMemberBase
	bitsNonMemberNext := nonMemberBase

	var result []trialPoint
	ticker := ticks{stepSize: 1.0 / granularity}
	prevItems := 0
	currentHashes := 0
	var qb *qbloom.Filter
	var bb *bitsbloom.BloomFilter

	for {
		numItems := ticker.next()
		load := float64(numItems) / float64(numBits)
		hashes := optimalHashes(numBits, numItems)

		if hashes != currentHashes {
			qb = qbloom.NewSeeded(numBits, hashes, 0)
			bb = bitsbloom.New(uint(numBits), uint(hashes))
			addRangeQBloom(qb, memberBase, uint64(numItems))
			addRangeBitsBloom(bb, memberBase, uint64(numItems))
			currentHashes = hashes
		} else {
			addRangeQBloom(qb, memberBase+uint64(prevItems), uint64(numItems-prevItems))
			addRangeBitsBloom(bb, memberBase+uint64(prevItems), uint64(numItems-prevItems))
		}

		qbloomFP, nextQBNonMember := falsePosRateAdaptiveQBloom(qb, qbNonMemberNext, probeBudget)
		bitsFP, nextBitsNonMember := falsePosRateAdaptiveBitsBloom(bb, bitsNonMemberNext, probeBudget)
		qbNonMemberNext = nextQBNonMember
		bitsNonMemberNext = nextBitsNonMember

		result = append(result, trialPoint{
			itemsPerBit: load,
			qbloomFP:    qbloomFP,
			bitsFP:      bitsFP,
		})

		prevItems = numItems
		if load >= maxItemsPerBit {
			break
		}
	}

	return result
}

func aggregateTrials(rows [][]trialPoint) ([]samplePoint, []samplePoint) {
	minLen := len(rows[0])
	for _, row := range rows[1:] {
		if len(row) < minLen {
			minLen = len(row)
		}
	}

	qbloomRows := make([]samplePoint, 0, minLen)
	bitsRows := make([]samplePoint, 0, minLen)
	for i := 0; i < minLen; i++ {
		qbTrials := make([]float64, 0, len(rows))
		bitsTrials := make([]float64, 0, len(rows))
		for _, row := range rows {
			qbTrials = append(qbTrials, row[i].qbloomFP)
			bitsTrials = append(bitsTrials, row[i].bitsFP)
		}
		qbloomRows = append(qbloomRows, summarizePoint(rows[0][i].itemsPerBit, qbTrials))
		bitsRows = append(bitsRows, summarizePoint(rows[0][i].itemsPerBit, bitsTrials))
	}
	return qbloomRows, bitsRows
}

func summarizePoint(itemsPerBit float64, trials []float64) samplePoint {
	avg := 0.0
	min := trials[0]
	max := trials[0]
	for _, trial := range trials {
		avg += trial
		if trial < min {
			min = trial
		}
		if trial > max {
			max = trial
		}
	}
	avg /= float64(len(trials))
	return samplePoint{itemsPerBit: itemsPerBit, avgFP: avg, minFP: min, maxFP: max}
}

func writeSeries(path string, rows []samplePoint) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatFloat(row.itemsPerBit, 'f', 8, 64),
			strconv.FormatFloat(row.avgFP, 'f', 8, 64),
			strconv.FormatFloat(row.minFP, 'f', 8, 64),
			strconv.FormatFloat(row.maxFP, 'f', 8, 64),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func falsePosRateAdaptiveQBloom(filter *qbloom.Filter, next uint64, probeBudget int) (float64, uint64) {
	total := 0
	falsePositives := 0
	var buf [8]byte
	for {
		binary.LittleEndian.PutUint64(buf[:], next)
		next++
		total++
		if filter.Contains(buf[:]) {
			falsePositives++
		}
		fpn := float64(falsePositives + 1)
		if math.Pow(fpn, 1.5)*float64(total) >= float64(probeBudget) {
			break
		}
	}
	return float64(falsePositives) / float64(total), next
}

func falsePosRateAdaptiveBitsBloom(filter *bitsbloom.BloomFilter, next uint64, probeBudget int) (float64, uint64) {
	total := 0
	falsePositives := 0
	var buf [8]byte
	for {
		binary.LittleEndian.PutUint64(buf[:], next)
		next++
		total++
		if filter.Test(buf[:]) {
			falsePositives++
		}
		fpn := float64(falsePositives + 1)
		if math.Pow(fpn, 1.5)*float64(total) >= float64(probeBudget) {
			break
		}
	}
	return float64(falsePositives) / float64(total), next
}

func addRangeQBloom(filter *qbloom.Filter, start uint64, count uint64) uint64 {
	var buf [8]byte
	for i := uint64(0); i < count; i++ {
		binary.LittleEndian.PutUint64(buf[:], start+i)
		filter.Add(buf[:])
	}
	return start + count
}

func addRangeBitsBloom(filter *bitsbloom.BloomFilter, start uint64, count uint64) uint64 {
	var buf [8]byte
	for i := uint64(0); i < count; i++ {
		binary.LittleEndian.PutUint64(buf[:], start+i)
		filter.Add(buf[:])
	}
	return start + count
}

func optimalHashes(numBits, expectedItems int) int {
	hashes := math.Round(math.Ln2 * float64(numBits) / float64(expectedItems))
	if hashes < 1 {
		return 1
	}
	return int(hashes)
}

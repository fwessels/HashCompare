/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"container/heap"
	"encoding/hex"
	"fmt"
	"github.com/aead/poly1305"
	"github.com/aead/siphash"
	"github.com/minio/blake2b-simd"
	"github.com/minio/highwayhash"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"
)

func Poly1305Sum(buf []byte, key [32]byte) []byte {
	h := poly1305.New(key)
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func Blake2bSum(buf []byte) []byte {
	h := blake2b.New512()
	h.Reset()
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func Blake2b256Sum(buf []byte) []byte {
	h := blake2b.New256()
	h.Reset()
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func SipHashSum(buf []byte, key [32]byte) []byte {
	h, _ := siphash.New128(key[0:16])
	h.Reset()
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func HighwayHash(buf []byte, key [32]byte) []byte {
	h, _ := highwayhash.New(key[:])
	h.Reset()
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func HighwayHash128(buf []byte, key [32]byte) []byte {
	h, _ := highwayhash.New128(key[:])
	h.Reset()
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func HighwayHash64(buf []byte, key [32]byte) []byte {
	h, _ := highwayhash.New64(key[:])
	h.Reset()
	h.Write(buf[:])
	sum := h.Sum(nil)
	return sum
}

func mask(b, m int) byte {
	s := uint(9 - m)
	return byte(((1 << s) - 1) << uint(b%m))
}

// Slice of sorted bytes
type SortBytes [][]byte

func (s SortBytes) Less(i, j int) bool {
	for index := range s[i] {
		if s[i][index] != s[j][index] {
			return s[i][index] < s[j][index]
		}
	}
	return false
}

func (s SortBytes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortBytes) Len() int {
	return len(s)
}

func TestHashPermutationsRange(msg []byte, key [32]byte, cpu int, cpuShift uint, algo string, results chan<- SortBytes) {

	var keys SortBytes

	for m := 8; m >= 1; m-- {

		bStart := (len(msg) * m * cpu) >> cpuShift
		bEnd := (len(msg) * m * (cpu + 1)) >> cpuShift

		for b := bStart; b < bEnd; b++ {
			// Change message with mask
			msg[b/m] = msg[b/m] ^ mask(b, m)

			var tag []byte
			switch algo {
			case "poly1305":
				tag = Poly1305Sum(msg, key)
			case "blake2b":
				tag = Blake2bSum(msg)
			case "blake2b-256":
				tag = Blake2b256Sum(msg)
			case "siphash":
				tag = SipHashSum(msg, key)
			case "highwayhash":
				tag = HighwayHash(msg, key)
			case "highwayhash128":
				tag = HighwayHash128(msg, key)
			case "highwayhash64":
				tag = HighwayHash64(msg, key)
			}

			keys = append(keys, tag)

			// Undo change
			msg[b/m] = msg[b/m] ^ mask(b, m)
		}
	}

	// sanity check if message not corrupted
	for i := range msg {
		if msg[i] != byte(i) {
			panic("memory corrupted")
		}
	}

	sort.Sort(keys)

	results <- keys
}

func TestHashPermutations(key [32]byte, msgSize uint, algo string) (permutations, zeroBits int, elapsed time.Duration) {
	start := time.Now()
	fmt.Println("Starting", start.Format("3:04PM"), "...")

	msg := make([]byte, msgSize)
	for i := range msg {
		msg[i] = byte(i)
	}

	cpus := runtime.NumCPU()
	// make sure nr of CPUs is a power of 2
	cpuShift := uint(len(strconv.FormatInt(int64(cpus), 2)) - 1)
	cpus = 1 << cpuShift

	var wg sync.WaitGroup
	results := make(chan SortBytes)

	for cpu := 0; cpu < cpus; cpu++ {

		wg.Add(1)
		go func(cpu int) {
			msgCopy := make([]byte, len(msg))
			copy(msgCopy, msg)
			defer wg.Done()
			TestHashPermutationsRange(msgCopy, key, cpu, cpuShift, algo, results)
		}(cpu)
	}

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results) // Close output channel
	}()

	keys := make([]SortBytes, 0, cpus)

	// Collect sub sorts
	for r := range results {
		keys = append(keys, r)
	}

	var tagprev *big.Int
	var smallest *big.Int

	h := &HexHeap{}
	heap.Init(h)

	// Push initial keys onto heap
	for stack, sb := range keys {
		if len(sb) > 0 {
			heap.Push(h, Hex{hex.EncodeToString(sb[0]), stack, 1})
		}
	}

	// Iterate over heap
	for h.Len() > 0 {
		hi := heap.Pop(h).(Hex)

		tag := new(big.Int)
		tag.SetString(hi.hex, 16)

		if tagprev != nil {
			diff := new(big.Int)
			diff.Sub(tag, tagprev)
			//fmt.Println(k, diff)

			if smallest == nil {
				smallest = new(big.Int)
				*smallest = *diff
			} else if smallest.Cmp(diff) > 0 {
				*smallest = *diff
			}
		}

		tagprev = tag

		if hi.index < len(keys[hi.stack]) { // Push new entry when still available
			heap.Push(h, Hex{hex.EncodeToString(keys[hi.stack][hi.index]), hi.stack, hi.index + 1})
		}
	}

	elapsed = time.Since(start)
	fmt.Println(smallest)
	return len(keys) * len(keys[0]), 8*len(keys[0][0]) - len(smallest.Text(2)), elapsed
}

// An HexHeap is a min-heap of hexadecimals.
type HexHeap []Hex

type Hex struct {
	hex   string
	stack int
	index int
}

func (h HexHeap) Len() int           { return len(h) }
func (h HexHeap) Less(i, j int) bool { return h[i].hex < h[j].hex }
func (h HexHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *HexHeap) Push(x interface{}) {
	*h = append(*h, x.(Hex))
}

func (h *HexHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func permuteAlgorithm(algo string) {
	rand.Seed(time.Now().UTC().UnixNano())

	var key [32]byte
	for i := range key {
		key[i] = byte(255 - i)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, '-', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "Permutations", "\t", "Zero bits", "\t", "Duration")

	const hundredFiftyMillionPermutations = 23
	for shift := uint(8); shift < hundredFiftyMillionPermutations; shift++ {
		permutations, zeroBits, elapsed := TestHashPermutations(key, 1<<shift, algo)
		fmt.Printf("Permutations: %d -- zero bits: %d -- duration: %v (%s)\n", permutations, zeroBits, elapsed, algo)
		fmt.Fprintln(w, permutations, "\t", zeroBits, "\t", elapsed)
	}
	fmt.Println()
	fmt.Println(algo)
	w.Flush()
}

func main() {
	
	//algo = "blake2b"
	//algo = "blake2b-256"
	//algo = "poly1305"
	//algo = "siphash"
	permuteAlgorithm("highwayhash")
	permuteAlgorithm("highwayhash128")
	//algo = "highwayhash128"
	//algo = "highwayhash64"
}

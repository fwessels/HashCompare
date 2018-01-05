/*
 * Minio Cloud Storage, (C) 2018 Minio, Inc.
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

package main_test

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"github.com/minio/highwayhash"
	sha256Avx512 "github.com/minio/sha256-simd"
	"golang.org/x/crypto/blake2b"
	"hash"
	"testing"
)

const size = 5 * 1024 * 1024

func benchmarkHashWithKey(b *testing.B, hash func(key []byte) (hash.Hash, error)) {
	var key [32]byte
	data := make([]byte, size)
	rand.Read(data)

	b.SetBytes(size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h, _ := hash(key[:])
		h.Write(data)
		h.Sum(nil)
	}
}

func benchmarkHash(b *testing.B, hash func() hash.Hash) {
	data := make([]byte, size)
	rand.Read(data)

	b.SetBytes(size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := hash()
		h.Write(data)
		h.Sum(nil)
	}
}

func BenchmarkHighwayHash(b *testing.B) {
	benchmarkHashWithKey(b, highwayhash.New)
}

func BenchmarkSHA256_AVX512(b *testing.B) {
	benchmarkAvx512(b, size)
}

func BenchmarkBlake2b(b *testing.B) {
	benchmarkHashWithKey(b, blake2b.New512)
}

func BenchmarkSHA1(b *testing.B) {
	benchmarkHash(b, sha1.New)
}

func BenchmarkMD5(b *testing.B) {
	benchmarkHash(b, md5.New)
}

func BenchmarkSHA512(b *testing.B) {
	benchmarkHash(b, sha512.New)
}

func BenchmarkSHA256(b *testing.B) {
	benchmarkHash(b, sha256.New)
}

// AVX512 code below

func benchmarkAvx512SingleCore(h512 []hash.Hash, body []byte) {

	for i := 0; i < len(h512); i++ {
		h512[i].Write(body)
	}
	for i := 0; i < len(h512); i++ {
		_ = h512[i].Sum([]byte{})
	}
}

func benchmarkAvx512(b *testing.B, size int) {

	server := sha256Avx512.NewAvx512Server()

	const tests = 16
	body := make([]byte, size)
	rand.Read(body)

	b.SetBytes(int64(len(body) * tests))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h512 := make([]hash.Hash, tests)
		for i := 0; i < tests; i++ {
			h512[i] = sha256Avx512.NewAvx512(server)
		}

		benchmarkAvx512SingleCore(h512, body)
	}
}

package main

import "fmt"
import "crypto/md5"

// url, gfy-url
func convert(input string, n, max int) []int {
	b := md5.Sum([]byte(input)) // 16 bytes
	result := make([]int, n)
	blocksize := (numbits(max) / 8) + 1
	j := 0
	for i := 0; i < n; i++ {
		result[i] = getInt(b[j:j+blocksize]) % max
		j += blocksize
	}
	return result
}

func getInt(b []byte) int {
	l := len(b)
	result := 0
	var shiftby uint32
	for i := 0; i < l; i++ {
		shiftby = uint32(8 * (l - i - 1))
		result |= int(b[i]) << shiftby
	}
	return result
}

func numbits(n int) int {
	result := 0
	for n > 0 {
		n = n / 2
		result++
	}
	return result
}

// max is implicitly 256
func convert2(input string, n int) []int {
	b := md5.Sum([]byte(input))
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = int(b[i])
	}
	return result
}

func main() {
	fmt.Println((numbits(32767) / 8) + 1)
	fmt.Println(convert("hivkdv", 8, 1024))

}

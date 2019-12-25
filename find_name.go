package main

import (
	"crypto/md5"
	"encoding/base64"

	"golang.org/x/text/width"

	//  "unicode/utf8"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"time"
)

type GoParam struct {
	Ss             string
	JapaneseCharss [][]string
}

// intToString はnを客たーコードとして文字列を返します
func intToString(n int) string {
	c := rune(n)
	return string(c)
}

// isJapanese は文字(1文字)が日本語かどうかを返します
func isJapanese(s string) bool {
	// Regular Expressions for Japanese Text
	// https://www.localizingjapan.com/blog/2012/01/20/regular-expressions-for-japanese-text/
	// Unicode code points regex: [\x3041-\x3096]
	// Unicode block property regex: \p{Hiragana}
	// Unicode code points regex: [\x30A0-\x30FF]
	// Unicode block property regex: \p{Katakana}
	// Unicode code points regex: [\x3400-\x4DB5\x4E00-\x9FCB\xF900-\xFA6A]
	// Unicode block property regex: \p{Han}
	r := regexp.MustCompile(`\p{Hiragana}|\p{Katakana}|\p{Han}`)
	return r.Match([]byte(s))
}

// isJapanese_ はisJapaneseと同様ですが、とりあえず使いません
func isJapanese_(s string) bool {
	p, _ := width.LookupString(s)
	return p.Kind() == width.EastAsianWide
}

// makeJapaneseChars はユニコードでU+0001～U+FFFFまでの日本語を文字の配列として返します
func makeJapaneseChars() []string {
	rs := []string{}
	for i := 1; i < 0x10000; i++ {
		s := intToString(i)
		if isJapanese(s) {
			rs = append(rs, s)
		}
	}
	return rs
}

// innerMakeStrings はmakeStringsの内部関数で、与えられたjapaneseChars配列の文字数分の文字列を生成し、評価関数で評価する
func innerMakeStrings(id int, jobs chan GoParam, finish chan<- bool, f func(string) bool, ss string, japaneseCharss [][]string) {
	count := len(japaneseCharss)
	if count > 0 {
		if count == 1 {
			// fmt.Printf("%d: %s\n", id, ss)
			for _, s := range japaneseCharss[0] {
				if !f(ss + s) {
					fmt.Printf("%d: finish<-\n", id)
					finish <- true
					return
				}
			}
		} else {
			// fmt.Printf("%d: (%s)\n", id, ss)
			js := japaneseCharss[1:]
			for _, s := range japaneseCharss[0] {
				jobs <- GoParam{ss + s, js}
			}
		}
	}
}

func worker(id int, f func(string) bool, jobs chan GoParam, finish chan<- bool, done <-chan bool) {
OUTER:
	for {
		select {
		default:
			fmt.Printf("%d: sleeping\n", id)
			time.Sleep(1 * time.Second)
		case _, ok := <-done:
			if !ok {
				fmt.Printf("%d: done\n", id)
				break OUTER
			}
		case job, ok := <-jobs:
			if ok {
				innerMakeStrings(id, jobs, finish, f, job.Ss, job.JapaneseCharss)
			} else {
				fmt.Printf("%d: break\n", id)
				break OUTER
			}
		}
	}
	fmt.Printf("%d: finish\n", id)
}

// processing は3文字までの名前に対して、文字列を評価する
func processing(japaneseChars []string, f func(string) bool) {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	jobs := make(chan GoParam, 65536)
	finish := make(chan bool)
	dones := []chan bool{}

	for i := 0; i < numCpu; i++ {
		done := make(chan bool, numCpu+1)
		go worker(i, f, jobs, finish, done)
		dones = append(dones, done)
	}
	// お仕事生成
	// 1文字
	jobs <- GoParam{"", [][]string{japaneseChars}}
	// 2文字
	jobs <- GoParam{"", [][]string{japaneseChars, japaneseChars}}
	// 3文字
	jobs <- GoParam{"", [][]string{japaneseChars, japaneseChars, japaneseChars}}

	// 終了を待つ
	<-finish
	for _, done := range dones {
		close(done)
	}
	fmt.Println("0: finish")
}

// calculateHash は文字列からbase64とMD5を求めます
func calculateHash(s string) (string, string) {
	// base64化する
	// echo "XXX" | base64 では最後に0x0aがついているので付加する
	b64 := base64.StdEncoding.EncodeToString([]byte(s + string(rune(0x0a))))
	// md5でハッシュ化
	// echo "XXX" | base64 | md5sumでは最後に0x0aがついているので付加する
	md5 := md5.New()
	io.WriteString(md5, b64+string(rune(0x0a)))
	c := md5.Sum(nil)
	r := fmt.Sprintf("%x", c)
	return b64, r
}

// process は文字列が答えかどうかを判断します
func process(s string) bool {
	_, r := calculateHash(s)
	if r == "d41d8cd98f00b204e9800998ecf8427e" {
		fmt.Println(s)
		return false
	}
	return true
}

// main はまいんちゃんです
func main() {
	japaneseChars := makeJapaneseChars()
	processing(japaneseChars, process)
}

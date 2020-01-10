package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"

	"golang.org/x/text/width"

	//  "unicode/utf8"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"sync"
	"time"
)

type Job struct {
	Ss             string
	JapaneseCharss [][]string
}

type Jobs struct {
	mutex       *sync.Mutex
	appendCount int
	popCount    int
	Jobs        []Job
}

func (j *Jobs) Append(job Job) {
	j.mutex.Lock()
	j.Jobs = append(j.Jobs, job)
	j.appendCount++
	j.mutex.Unlock()
}

func (j *Jobs) Pop() (Job, error) {
	j.mutex.Lock()
	if len(j.Jobs) <= 0 {
		j.mutex.Unlock()
		err := errors.New("ジョブが空です")
		return Job{}, err
	}
	job := j.Jobs[0]
	j.Jobs = j.Jobs[1:]
	j.popCount++
	j.mutex.Unlock()
	return job, nil
}

// intToString はnをキャラクターコードとして文字列を返します
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
func innerMakeStrings(id int, jobs *Jobs, finish chan<- bool, f func(string) bool, ss string, japaneseCharss [][]string) {
	count := len(japaneseCharss)
	if count > 0 {
		if count == 1 {
//			fmt.printf("%d: %s\n", id, ss)
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
			if ss == "" {
				for _, s := range japaneseCharss[0] {
					jobs.Append(Job{ss + s, js})
				}
			} else {
				for _, s := range japaneseCharss[0] {
					innerMakeStrings(id, jobs, finish, f, ss + s, js)
				}
			}
		}
	}
}

func worker(id int, f func(string) bool, jobs *Jobs, finish chan<- bool, done <-chan bool) {
OUTER:
	for {
		select {
		default:
			job, err := jobs.Pop()
			if err == nil {
				innerMakeStrings(id, jobs, finish, f, job.Ss, job.JapaneseCharss)
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		case _, ok := <-done:
			if !ok {
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

	jobs := Jobs{new(sync.Mutex), 0, 0, []Job{}}
	dones := []chan bool{}
	finish := make(chan bool, numCpu + 1)

	for i := 0; i < numCpu; i++ {
		done := make(chan bool, numCpu+1)
		go worker(i, f, &jobs, finish, done)
		dones = append(dones, done)
	}
	// お仕事生成
	id := 0
	// 1文字
	innerMakeStrings(id, &jobs, finish, f, "", [][]string{japaneseChars})
	// 2文字
	innerMakeStrings(id, &jobs, finish, f, "", [][]string{japaneseChars, japaneseChars})
	// 3文字
	innerMakeStrings(id, &jobs, finish, f, "", [][]string{japaneseChars, japaneseChars, japaneseChars})

	// 終了を待つ
	<-finish
	for _, done := range dones {
		close(done)
	}
	fmt.Printf("%d: finish, AppendCount = %d, PopCount = %d\n", id, jobs.appendCount, jobs.popCount)
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

// main はまいんちゃんです
func main() {
  count := 0
  f := func (s string) bool {
    count += 1
    //    fmt.Printf("%d, %s\n", count, s)
	  _, r := calculateHash(s)
	  if r == "d41d8cd98f00b204e9800998ecf8427e" {
		  fmt.Println(s)
		  return false
	  }
    return true
  }
	japaneseChars := makeJapaneseChars()
  processing(japaneseChars, f)
}

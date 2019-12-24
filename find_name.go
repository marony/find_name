package main

import (
	"crypto/md5"
	"encoding/base64"

	"golang.org/x/text/width"

	//  "unicode/utf8"
	"fmt"
	"io"
	"regexp"
)

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
func innerMakeStrings(ss string, f func(string) bool, japaneseCharss ...[]string) bool {
	count := len(japaneseCharss)
	if count <= 0 {
		return false
	} else {
		if count == 1 {
			fmt.Println(ss)
			for _, s := range japaneseCharss[0] {
				if !f(ss + s) {
					return false
				}
			}
		} else {
			js := japaneseCharss[1:]
			for _, s := range japaneseCharss[0] {
				r := innerMakeStrings(ss+s, f, js...)
				if !r {
					return false
				}
			}
		}
	}
	return true
}

// makeStrings は3文字までの名前に対して、文字列を評価する
func makeStrings(japaneseChars []string, f func(string) bool) bool {
	// 1文字目
	if !innerMakeStrings("", f, [][]string{japaneseChars}...) {
		return false
	}
	// 2文字目
	if !innerMakeStrings("", f, [][]string{japaneseChars, japaneseChars}...) {
		return false
	}
	// 3文字目
	if !innerMakeStrings("", f, [][]string{japaneseChars, japaneseChars, japaneseChars}...) {
		return false
	}
	return true
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
	//  makeStrings(func (s string) { fmt.Println(s) })
	makeStrings(japaneseChars, process)

	msg := "Hello, 世界"
	encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	fmt.Println(encoded)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}
	fmt.Println(string(decoded))
}

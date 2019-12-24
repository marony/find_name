package main

import (
  "testing"
)

func TestCalculateHash(t *testing.T) {
  {
    b64, r := calculateHash("篤")
    if b64 != "56+kCg==" {
      t.Error("base64エラー")
    }
    if r != "b3d949edf1f245015e8e44590ff114ef" {
      t.Error("MD5エラー")
    }
  }
  {
    b64, r := calculateHash("篤志")
    if b64 != "56+k5b+XCg==" {
      t.Error("base64エラー")
    }
    if r != "c7a61623ac1ed1b89b633fd98072c13f" {
      t.Error("MD5エラー")
    }
  }
}

func TestEnumeration(t *testing.T) {
  japaneseChars := makeJapaneseChars()
  {
    success := false
    f := func (s string) bool {
      if s == "篤" {
        success = true
        return false
      }
      return true
    }
    makeStrings(japaneseChars, f)
    if !success {
      t.Error("失敗")
    }
  }
  {
    success := false
    f := func (s string) bool {
      if s == "篤志" {
        success = true
        return false
      }
      return true
    }
    makeStrings(japaneseChars, f)
    if !success {
      t.Error("失敗")
    }
  }
}


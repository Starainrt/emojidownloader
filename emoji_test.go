package main

import (
	"fmt"
	"testing"
)

func Test_ParseEmojiFromUrl(t *testing.T) {
	url := "https://b612.icu"
	emo := NewEmojis()
	err := emo.LoadAndParse(url, true)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	if len(emo.counts)==0 {
		t.FailNow()
	}
	fmt.Println(emo.counts)
	fmt.Println(emo.allCategories)
}

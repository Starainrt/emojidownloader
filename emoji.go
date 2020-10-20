package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"b612.me/staros"

	"b612.me/starnet"
)

type TarFile struct {
	gw       *gzip.Writer
	tw       *tar.Writer
	destPath string
	fw       *os.File
	mu       sync.Mutex
}

type Emoji struct {
	ShortCode string `json:"shortcode"`
	Url       string `json:"url"`
	StaticUrl string `json:"static_url"`
	Picker    bool   `json:"visible_in_picker"`
	Category  string `json:"category"`
}

type EmojiDownload struct {
	EmojiList        []Emoji
	Category         []string
	SaveFolder       string
	Proxy            string
	ShortCodeRegexp  string
	Zipped           bool
	DeletedZipOrigin bool
	ReplaceCodeOld   string
	ReplaceCodeNew   string
	Thread           int
}

func Download(in EmojiDownload) {
	var mu *sync.WaitGroup = &sync.WaitGroup{}
	var emojiList []string
	zipMap := make(map[string]*TarFile)
	threadChan := make(chan struct{}, in.Thread)
	if in.ShortCodeRegexp != "" {
		_, err := regexp.Compile(in.ShortCodeRegexp)
		if err != nil {
			fmt.Println("Cannot Parse Reg String,Please Check!", err)
			return
		}
	}
	if in.ReplaceCodeOld != "" {
		_, err := regexp.Compile(in.ReplaceCodeOld)
		if err != nil {
			fmt.Println("Cannot Parse Reg String,Please Check!", err)
			return
		}
	}
	for _, cat := range in.Category {
		if cat == "" {
			cat = "未分类"
		}
		folder := filepath.Join(in.SaveFolder, cat)
		if !staros.Exists(folder) {
			err := os.MkdirAll(folder, 0755)
			if err != nil {
				fmt.Println("Cannot Create Folder", err)
				return
			}
		}
		if in.Zipped {
			var err error
			zipMap[cat], err = NewTar(filepath.Join(in.SaveFolder, cat+".tar.gz"))
			if err != nil {
				fmt.Println("Cannot Create gzip File", err)
				return
			}
		}
	}

	for _, emoji := range in.EmojiList {
		for _, cat := range in.Category {
			if emoji.Category == cat {
				if !shouldDownload(emoji.ShortCode, in.ShortCodeRegexp) {
					break
				}
				if cat == "" {
					emoji.Category = "未分类"
				}
				threadChan <- struct{}{}
				mu.Add(1)
				go func(threadChan chan struct{}, mu *sync.WaitGroup, emoji Emoji) {
					defer mu.Done()
					defer func() {
						<-threadChan
					}()
					data, err := doDownloadUrl(emoji.Url, in.Proxy)
					if err != nil {
						fmt.Printf("Cannot Download %s,Url is %s and Reason is %v", emoji.ShortCode, emoji.Url, err)
						return
					}
					path := filepath.Join(in.SaveFolder, emoji.Category, newShortCode(emoji.ShortCode, in.ReplaceCodeOld, in.ReplaceCodeNew)+path.Ext(emoji.Url))
					err = ioutil.WriteFile(path, data, 0755)
					if err != nil {
						fmt.Printf("Cannot Write Data into Disk %s,Url is %s and Reason is %v\n", emoji.ShortCode, emoji.Url, err)
						return
					}
					if in.Zipped {
						err := zipMap[emoji.Category].AddFile(path, filepath.Base(path))
						if err != nil {
							fmt.Println("Cannot Add data to gzip", err)
						}
					}
					if in.DeletedZipOrigin {
						emojiList = append(emojiList, path)
					}
					fmt.Printf("Download %s Succeed\n", emoji.ShortCode)
				}(threadChan, mu, emoji)
				break
			}
		}
	}
	mu.Wait()
	if in.Zipped {
		for _, v := range zipMap {
			v.Finish()
		}
	}
	if in.DeletedZipOrigin {
		for _, v := range emojiList {
			os.Remove(v)
		}
	}
}

func newShortCode(name, old, new string) string {
	if old == "" {
		return name
	}
	return regexp.MustCompile(old).ReplaceAllString(name, new)
}
func shouldDownload(name, check string) bool {
	if check == "" {
		return true
	}
	match, _ := regexp.MatchString(check, name)
	return match
}

func AnalyseEmojis(data []byte) ([]Emoji, error) {
	var result []Emoji
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func SumEmojiCats(data []Emoji) map[string]int {
	res := make(map[string]int)
	for _, v := range data {
		res[v.Category]++
	}
	return res
}

func doDownloadUrl(url, proxy string) ([]byte, error) {
	req := starnet.NewRequests(url, nil, "GET")
	if proxy != "" {
		req.Proxy = proxy
	}
	data, err := starnet.Curl(req)
	return data.RecvData, err
}

func NewTar(dstPath string) (*TarFile, error) {
	var res TarFile
	var err error
	res.fw, err = os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	res.gw = gzip.NewWriter(res.fw)
	res.tw = tar.NewWriter(res.gw)
	res.destPath = dstPath
	return &res, nil
}

func (tf *TarFile) AddFile(path, name string) error {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	fso, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fso.Close()
	stats, err := fso.Stat()
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(stats, "")
	if err != nil {
		return err
	}
	hdr.Name = name
	err = tf.tw.WriteHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(tf.tw, fso)
	return err
}

func (tf *TarFile) Finish() error {
	err := tf.tw.Close()
	if err != nil {
		return err
	}
	err = tf.gw.Close()
	if err != nil {
		return err
	}
	return tf.fw.Close()
}

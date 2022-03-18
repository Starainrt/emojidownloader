package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"b612.me/stario"
	"b612.me/starnet"
	"b612.me/staros"
)

type Emoji struct {
	ShortCode string `json:"shortcode"`
	Url       string `json:"url"`
	StaticUrl string `json:"static_url"`
	Picker    bool   `json:"visible_in_picker"`
	Category  string `json:"category"`
}

type EmojiOpt func(*EmojiOpts)

type EmojiOpts struct {
	Zip2Tarfile        bool
	DeletedOriginIfZip bool
	IgnoreErr          bool
	AuthCookie         string
	filterRp           *regexp.Regexp
	rpCodeOld          *regexp.Regexp
	rpNew              string
	Threads            int
	SaveFolders        string
	Proxy              string
}

type Emojis struct {
	Lists           []Emoji
	AllowCategories map[string]bool
	zipMaps         map[string]*TarFile
	OriginJson      string
	counts          map[string]int
	allCategories   []string
	EmojiOpts
}

func (e *Emojis) Parse() error {
	err := json.Unmarshal([]byte(e.OriginJson), &e.Lists)
	if err != nil {
		return err
	}
	e.allCategories = []string{}
	e.counts = make(map[string]int)
	for _, v := range e.Lists {
		if v.Category == "" {
			v.Category = "未分类"
		}
		if _, ok := e.counts[v.Category]; !ok {
			e.allCategories = append(e.allCategories, v.Category)
		}
		e.counts[v.Category] = e.counts[v.Category] + 1
	}
	return nil
}

func (e *Emojis) replaceName(name string) string {
	if e.rpCodeOld == nil {
		return name
	}
	return e.rpCodeOld.ReplaceAllString(name, e.rpNew)
}

func (e *Emojis) filterName(name string) bool {
	if e.filterRp == nil {
		return true
	}
	return e.filterRp.MatchString(name)
}

func (e *Emojis) generateRequest(url string, data []byte, nettype string) starnet.Request {
	req := starnet.NewRequests(url, nil, "GET", starnet.WithProxy(e.Proxy), starnet.WithUserAgent("b612 emoji downloader "+version))
	if e.AuthCookie != "" {
		req.AddSimpleCookie("_session_id", e.AuthCookie)
	}
	return req
}

func (e *Emojis) Download(fns ...func(v Emoji, finished bool)) error {
	SetUmask(0)
	defer UnsetUmask()
	var fn func(v Emoji, finished bool) = nil
	if len(fns) != 0 {
		fn = fns[0]
	}
	req := e.generateRequest("", nil, "GET")
	download := func(d Emoji) error {
		if fn != nil {
			fn(d, false)
		}
		savePath := filepath.Join(e.SaveFolders, d.Category)
		if !staros.Exists(d.Category) {
			err := os.MkdirAll(savePath, 0755)
			if err != nil {
				return err
			}
		}
		req.Url = d.Url
		data, err := starnet.Curl(req)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(savePath, e.replaceName(d.ShortCode)+path.Ext(d.Url))
		err = ioutil.WriteFile(dstPath, data.RecvData, 0644)
		if e.Zip2Tarfile {
			var err error
			if _, ok := e.zipMaps[d.Category]; !ok {
				e.zipMaps[d.Category], err = NewTar(filepath.Join(e.SaveFolders, d.Category+".tar.gz"))
				if err != nil {
					delete(e.zipMaps, d.Category)
					return err
				}
			}
			e.zipMaps[d.Category].AddFile(dstPath, filepath.Base(dstPath))
		}
		if fn != nil {
			fn(d, true)
		}
		return err
	}
	wg := stario.NewWaitGroup(e.Threads)
	errChan := make(chan error, 1)
	var err error
exitfor:
	for _, v := range e.Lists {
		if len(e.AllowCategories) == 0 || e.AllowCategories[v.Category] {
			if !e.filterName(v.ShortCode) {
				continue
			}
			t := v
			select {
			case err = <-errChan:
				break exitfor
			default:
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err == nil {
					if err := download(t); err != nil && !e.IgnoreErr {
						errChan <- err
					}
				}
			}()
		}
	}
	wg.Wait()
	for k, v := range e.zipMaps {
		v.Finish()
		if k != "" && e.DeletedOriginIfZip {
			os.RemoveAll(filepath.Join(e.SaveFolders, k))
		}
	}
	return err
}

func (e *Emojis) Load(fpath string, isSite bool) error {
	if isSite {
		fpath += "/api/v1/custom_emojis"
		data, err := starnet.Curl(e.generateRequest(fpath, nil, "GET"))
		if err != nil {
			return err
		}
		if data.RespHttpCode != 200 {
			return fmt.Errorf("http code not correct! got %v,resp text is %v", data.RespHttpCode, string(data.RecvData))
		}
		e.OriginJson = string(data.RecvData)
		return nil
	}
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	e.OriginJson = string(data)
	return nil
}

func (e *Emojis) LoadAndParse(fpath string, isSite bool) error {
	if err := e.Load(fpath, isSite); err != nil {
		return err
	}
	return e.Parse()
}

func (e *Emojis) CategoryCount() map[string]int {
	if e.counts == nil {
		return map[string]int{}
	}
	return e.counts
}

func (e *Emojis) SetDownloadCategories(cates ...string) {
	if e.AllowCategories == nil {
		e.AllowCategories = make(map[string]bool)
	}
	for _, v := range cates {
		e.AllowCategories[v] = true
	}
}

func (e *Emojis) SetFilter(filter string) error {
	if filter == "" {
		e.filterRp = nil
		return nil
	}
	regp, err := regexp.Compile(filter)
	if err != nil {
		return err
	}
	e.filterRp = regp
	return nil
}

func (e *Emojis) SetReplace(rp string, rpnew string) error {
	e.rpNew = rpnew
	if rp == "" {
		e.rpCodeOld = nil
		return nil
	}
	regp, err := regexp.Compile(rp)
	if err != nil {
		return err
	}
	e.rpCodeOld = regp
	return nil
}

func (e *Emojis) Counts() (int, []string, map[string]int) {
	return len(e.Lists), e.allCategories, e.counts
}

func NewEmojis(opt ...EmojiOpt) *Emojis {
	defaultEmojis := &Emojis{
		AllowCategories: make(map[string]bool),
		zipMaps:         make(map[string]*TarFile),
		counts:          make(map[string]int),
		EmojiOpts: EmojiOpts{
			Threads:     1,
			SaveFolders: "./myEmojis",
		},
	}
	for _, v := range opt {
		v(&defaultEmojis.EmojiOpts)
	}
	return defaultEmojis
}

func WithAuthCookie(cookieValue string) EmojiOpt {
	return func(eo *EmojiOpts) {
		eo.AuthCookie = cookieValue
	}
}

func WithProxy(p string) EmojiOpt {
	return func(eo *EmojiOpts) {
		eo.Proxy = p
	}
}

func WithZipFile(z bool) EmojiOpt {
	return func(eo *EmojiOpts) {
		eo.Zip2Tarfile = z
	}
}

func WithSaveFolder(f string) EmojiOpt {
	return func(eo *EmojiOpts) {
		eo.SaveFolders = f
	}
}

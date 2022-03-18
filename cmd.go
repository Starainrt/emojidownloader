package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"b612.me/starlog"
	"github.com/spf13/cobra"
)

var emo = NewEmojis()
var regOld, regNew, regFilter string
var allowCate []string
var useJson bool

func init() {
	cmdRoot.AddCommand(cmdMain, cmdShow)
	cmdMain.Flags().StringVarP(&emo.SaveFolders, "savepath", "s", "./myEmojis", "emoji下载文件夹")
	cmdMain.Flags().StringVarP(&emo.AuthCookie, "cookie", "c", "", "认证cookie,填入_session_id的值")
	cmdMain.Flags().BoolVarP(&emo.IgnoreErr, "ignore-error", "i", false, "忽略下载错误")
	cmdMain.Flags().BoolVarP(&emo.Zip2Tarfile, "zip", "z", true, "压缩为tar.gz文件")
	cmdMain.Flags().BoolVarP(&emo.DeletedOriginIfZip, "delete-after-zip", "d", false, "压缩为tar.gz文件后删除原始文件")
	cmdMain.Flags().StringSliceVarP(&allowCate, "allow-download-category", "a", []string{}, "要下载的分类")
	cmdMain.Flags().StringVarP(&regFilter, "filter", "f", "", "emoji名称白名单正则表达式")
	cmdMain.Flags().StringVarP(&regOld, "replace-old", "o", "", "emoji名称替旧名称正则表达式")
	cmdMain.Flags().StringVarP(&regNew, "replace-new", "r", "", "emoji名称替换新字符串")
	cmdMain.Flags().IntVarP(&emo.Threads, "threads", "n", 4, "同时下载协程数")
	cmdMain.Flags().StringVarP(&emo.Proxy, "proxy", "p", "", "使用代理")
	cmdRoot.PersistentFlags().BoolVarP(&useJson, "use-json-file", "j", false, "使用json文件而不是url")
}

var cmdRoot = &cobra.Command{
	Use:   "get",
	Short: "Mastodon Emoji Downloader",
}

var cmdMain = &cobra.Command{
	Use:   "get",
	Short: "Mastodon Emoji Downloader",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			starlog.Errorln("请输入一个url或文件地址")
			os.Exit(1)
		}
		if regFilter != "" {
			rpgF, err := regexp.Compile(regFilter)
			if err != nil {
				starlog.Errorln("invalid regexp:", regFilter, err)
			}
			emo.filterRp = rpgF
		}
		if regOld != "" {
			rpgF, err := regexp.Compile(regOld)
			if err != nil {
				starlog.Errorln("invalid regexp:", regOld, err)
			}
			emo.rpCodeOld = rpgF
			emo.rpNew = regNew
		}
		if emo.Threads <= 0 {
			starlog.Errorln("并发下载数不能小于等于0", emo.Threads)
			os.Exit(3)
		}
		url := strings.ToLower(strings.TrimSpace(args[0]))
		if !useJson && strings.Index(url, "https://") != 0 {
			url = "https://" + url
		}
		err := emo.LoadAndParse(url, !useJson)
		if err != nil {
			starlog.Errorln("load emoji lists failed:", err)
			os.Exit(4)
		}
		if len(allowCate) != 0 {
			myMap := emo.CategoryCount()
			for _, v := range allowCate {
				if _, ok := myMap[v]; !ok {
					starlog.Errorln(v, "not in the categroy lists,please check")
					os.Exit(5)
				}
				emo.AllowCategories[v] = true
			}
		}
		err = emo.Download(showFn)
		if err != nil {
			starlog.Errorln("download failed", err)
		}
		starlog.Infoln("finished")
	},
}

var cmdShow = &cobra.Command{
	Use:   "category",
	Args:  cobra.ExactArgs(1),
	Short: "Show Mastodon Emoji Categories",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			starlog.Errorln("请输入一个url或文件地址")
			os.Exit(1)
		}
		url := strings.ToLower(strings.TrimSpace(args[0]))
		if !useJson && strings.Index(url, "https://") != 0 {
			url = "https://" + url
		}
		err := emo.LoadAndParse(url, !useJson)
		if err != nil {
			starlog.Errorln("load emoji lists failed:", err)
			os.Exit(4)
		}
		ct, orderSlice, allCat := emo.Counts()
		fmt.Println("按分类解析结果如下：")
		fmt.Printf("%-5s %-10s %-28s\n", "序号", "表情个数", "分类名")
		for k, v := range orderSlice {
			fmt.Printf("%-5v %-10d %-28s\n", k+1, allCat[v], v)
		}
		starlog.Green("在%d个分类中共找到%d个表情\n", len(orderSlice), ct)
		starlog.Infoln("已完成，保存到", emo.SaveFolders)
	},
}

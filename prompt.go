package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"b612.me/stario"
	"b612.me/starlog"
)

var showFn = func(v Emoji, finished bool, err error) {
	if err != nil {
		starlog.Errorf("下载出错 分类:%-10s emoji:%s error:%v\n", v.Category, v.ShortCode, err)
		return
	}
	if !finished {
		starlog.Noticef("开始下载 分类:%-10s emoji:%s \n", v.Category, v.ShortCode)
		return
	}
	starlog.Infof("下载完成 分类:%-10s emoji:%s \n", v.Category, v.ShortCode)
}

func PromotMode() {
	bunny()
	starlog.SetLevelColor(starlog.LvNotice, []starlog.Attr{starlog.FgHiYellow})
	emo := NewEmojis()
	checkFn := func(fn func(emo *Emojis) int) {
		if code := fn(emo); code != 0 {
			stario.StopUntil("出现错误，按任意键结束", "", true)
			os.Exit(code)
		}
	}
	for _, v := range []func(*Emojis) int{plParseJson, plGetDownloadChoice, plRegexp, plDownload} {
		checkFn(v)
	}
	stario.StopUntil("下载完毕，按任意键退出", "", true)
}

func plParseJson(emo *Emojis) int {
	var fromSite bool
	var fpath string
	fmt.Println("mastodon emoji downloader " + VERSION)
	fmt.Println("当前为交互模式，如果您想使用命令行模式，请执行--help查看用法\n\n")
	fmt.Println("今天您想要来点什么？")
	fmt.Println("1.从url下载mastodon emoji\n")
	fmt.Println("2.从json下载mastodon emoji\n")
	for {
		choice := stario.MessageBox("请输入您的选择：", "0").MustInt()
		if choice != 1 && choice != 2 {
			starlog.Red("请输入1或者2哦，您的输入为%d\n", choice)
			continue
		}
		if choice == 1 {
			fromSite = true
		}
		break
	}
	for {
		if fromSite {
			fmt.Print("请输入Mastodon域名,不带https：")
		} else {
			fmt.Print("请输入Json文件地址：")
		}
		fpath = stario.MessageBox("", "").MustString()
		if fpath == "" {
			starlog.Red("请实际输入哦")
			continue
		}
		break
	}
	if fromSite {
		if strings.Index(fpath, "https://") != 0 {
			fpath = "https://" + fpath
		}
		if stario.YesNo("是否使用代理？(y/N)", false) {
			emo.Proxy = stario.MessageBox("请输入代理地址:", "").MustString()
		}
		if stario.YesNo("是否设置Mastodon Cookie？(y/N)", false) {
			emo.AuthCookie = stario.MessageBox("请输入_session_id Cookie的值:", "").MustString()
		}
	}
	starlog.Infoln("解析中......")
	err := emo.LoadAndParse(fpath, fromSite)
	if err != nil {
		starlog.Errorln("解析失败！请检查", err)
		return 1
	}
	return 0
}

func plGetDownloadChoice(emo *Emojis) int {
	ct, orderSlice, allCat := emo.Counts()
	fmt.Println("按分类解析结果如下：")
	fmt.Printf("%-5s %-10s %-28s\n", "序号", "表情个数", "分类名")
	for k, v := range orderSlice {
		fmt.Printf("%-5v %-10d %-28s\n", k+1, allCat[v], v)
	}
	starlog.Green("在%d个分类中共找到%d个表情\n", len(orderSlice), ct)
exitfor:
	for {
		choice, err := stario.MessageBox("请输入您要下载的分类序号，如需下载多个分类，用英文逗号分隔多个序号，下载全部表情直接回车:", "0").SliceInt(",")
		if err != nil {
			starlog.Errorln("您的输入有误，请输入数字，用英文逗号分隔，请检查后重新输入", err)
			continue
		}
		for _, v := range choice {
			if v == 0 {
				emo.AllowCategories = make(map[string]bool)
				fmt.Println("准备下载：全部分类")
				break exitfor
			}
			emo.AllowCategories[orderSlice[v-1]] = true
			fmt.Println("准备下载：", orderSlice[v-1])
		}
		break
	}
	return 0
}

func plRegexp(emo *Emojis) int {
	if stario.YesNo("是否开启白名单emoji名称？(y/N)", false) {
		for {
			rgpR, err := regexp.Compile(stario.MessageBox("请输入白名单正则表达式:", "").MustString())
			if err != nil {
				starlog.Errorln("正则表达式不正确", err)
				continue
			}
			emo.filterRp = rgpR
			break
		}
	}
	if stario.YesNo("是否替换emoji名称？(y/N)", false) {
		for {
			rgpR, err := regexp.Compile(stario.MessageBox("请输入匹配emoji正则表达式:", "").MustString())
			if err != nil {
				starlog.Errorln("正则表达式不正确", err)
				continue
			}
			emo.rpCodeOld = rgpR
			break
		}
		emo.rpNew = stario.MessageBox("请输入替换后的新名称：", "").MustString()
	}
	return 0
}

func plDownload(emo *Emojis) int {
	emo.SaveFolders = stario.MessageBox("请输入保存文件夹(默认：./myEmojis)：", "./myEmojis").MustString()
	emo.IgnoreErr = stario.YesNo("是否忽略下载中的单个emoji下载错误？(Y/n)", true)
	emo.Zip2Tarfile = stario.YesNo("是否打包为压缩文件？(Y/n)", true)
	if emo.Zip2Tarfile {
		emo.DeletedOriginIfZip = stario.YesNo("是否在打包为压缩文件后删除原始的下载文件(Y/n)", true)
	}
	for {
		emo.Threads = stario.MessageBox("并发下载量(默认：16)：", "16").MustInt()
		if emo.Threads <= 0 {
			starlog.Red("输入非法", emo.Threads)
			continue
		}
		break
	}
	var fn func(v Emoji, finished bool, err error) = nil
	if stario.YesNo("是否显示下载日志？(Y/n)", true) {
		fn = showFn
	}
	starlog.Infoln("开始下载...")
	err := emo.Download(fn)
	if err != nil {
		starlog.Errorln("下载失败：", err)
		return 1
	}
	starlog.Infoln("下载成功!保存到", emo.SaveFolders)
	fmt.Println(`tips:一键导入表情：
非docker：
RAILS_ENV=production bin/tootctl emoji import --category=自定义表情分类名 表情tar.gz文件地址

docker：
docker cp ./表情地址 web服务docker名:/tmp/表情名.tar.gz
docker-compose exec web bin/tootctl emoji import --category=自定义表情分类名 /tmp/表情名.tar.gz`)
	return 0
}

func bunny() {
	str := `
                                                            
                                                            
                  WMMMMa             rMMMMM:                
                 MM   ,MMi          MM7   rM.               
                 M      7MX       ;MZ      Mr               
                 MB      .MX     rM;      7M                
                 .M       :M:   iM:       M;                
                  aM       SM   M7       MZ                 
                   WM       M@ M0       M8                  
                    WM       MMM      ,MS                   
                     0M,     XM      SMi                    
                 ;aW@BMZ             MMBW0S:                
              aMM@Z;                    ,70MMW7             
           7MM0;                             X@MB,          
         XMM7                                   ZMM,        
       ,MM:      .        ;,     ;:      ,,,.     2MW       
      aMS    .ii;i;ii.   XMM ii:,MM   .i;i;i;ii.    MM,     
     BM     :;ii:i:ii;:   ;..MMM ii  .;;:i:i:ii;,    SMi    
    ZM     .;ii:i:i:iir       8      :;:i:::::ii;     rM    
    M.     .;i:i:i:i:i;       0      .;i:i:i:::;i      0M   
   MM       :;;iiiii;;,       W       ,i;iiiii;:        Mi  
   M;         ,iiiii,        .B         .::::,          M@  
   M:                         B                         MM  
   MZ                         0                         MS  
   rM  .S ;Xaaa0@MMMMMMMMMMMMMMMMMMMMMMMMMMMM0ZZZX,.2  ;M   
    MM ;8M0Z8MMMMMMWMa;7r7rr;;i;;r;rr7;MM8BW0MMMZ8MB8  Mr   
     MM rM    0M    Ma.:iiii;i;i;i;ii:.MX   ,Mr   .M iMS    
      aMrM8    SM;  XM.ii;;r;r;r;;;;ii:M   BM.    M88M:     
        8MMa    :MB  M0.;;r;;;;;;;;;i,MZ  MM     MMMS       
          ;MW     MM  MW:ii;;;;;i;i:;MW 7M8     MM.         
            MM     SM; BM0Xii:iii;2WMS BM,    rM8           
             8M8         aWM@@W@@MBX         MMi            
              .MM,           ,:.           7M8                     												   
													  
`
	fmt.Println(str)
}

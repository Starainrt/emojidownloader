package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"b612.me/starainrt"
)

var fetchEmojiLists []Emoji

func main() {
	Usagi()
	fmt.Println("今天要来点兔子吗？\n")
	proxy := starainrt.MessageBox("Step 1. 设置下载代理（如:http://127.0.0.1:8080），若无需设置代理请直接回车：", "")
	fmt.Println("\nStep 2. 选择实例表情解析方式 \n1.输入实例URL直接下载解析表情\n2.自行获取实例表情包列表，并在此处理（实例开启authorized_fetch选则此项）\n")
	chose := starainrt.MsgBox("请输入你的选择：", "1")
	switch chose {
	case "1":
		err := Chose01(proxy)
		if err != nil {
			fmt.Println(err)
			Exit()
		}
	default:
		err := Chose02()
		if err != nil {
			fmt.Println(err)
			Exit()
		}
	}
	sumData := SumEmojiCats(fetchEmojiLists)
	var choseLists []string
	fmt.Println("\n")
	i := 0
	for k, v := range sumData {
		i++
		p := k
		if k == "" {
			p = "未分类"
		}
		fmt.Println(i, p, "共"+strconv.Itoa(v)+"个表情")
		choseLists = append(choseLists, k)
	}
	inputOriData := starainrt.MessageBox("\n请选择需要下载的分组序号，逗号分割：", "1")
	tmpSlice := strings.Split(inputOriData, ",")
	var myChose []string
	for _, v := range tmpSlice {
		tmp, _ := strconv.Atoi(v)
		tmp--
		if tmp < 0 {
			continue
		}
		if tmp >= len(choseLists) {
			continue
		}
		myChose = append(myChose, choseLists[tmp])
	}
	fmt.Print("\n将要下载：")
	for _, v := range myChose {
		if v == "" {
			fmt.Print("未分类 ")
		}
		fmt.Print(v + " ")
	}
	fmt.Print("\n\n")
	var downloadPara EmojiDownload
	downloadPara.EmojiList = fetchEmojiLists
	downloadPara.Category = myChose
	downloadPara.Proxy = proxy
	downloadPara.SaveFolder = "./mastemojis"
	downloadPara.Zipped = starainrt.YesNo("是否压缩为tar.gz文件？（yes/no）:", true)
	if downloadPara.Zipped {
		downloadPara.DeletedZipOrigin = starainrt.YesNo("压缩后是否删除源文件？（yes/no）:", true)
	}
	downloadPara.ShortCodeRegexp = starainrt.MessageBox("shortCode过滤条件（正则）：", "")
	downloadPara.ReplaceCodeOld = starainrt.MessageBox("shortCode替换条件旧字符串（正则）：", "")
	downloadPara.ReplaceCodeNew = starainrt.MessageBox("shortCode替换条件新字符串：", "")
	thread := starainrt.MsgBox("异步下载协程数(默认：4)：", "4")
	threadInt, _ := strconv.Atoi(thread)
	if threadInt <= 0 {
		threadInt = 1
	}
	downloadPara.Thread = threadInt
	fmt.Println("处理中，请等待……")
	Download(downloadPara)
	fmt.Println()
	fmt.Println(`完成，请自行检查，象中输入下面的命令一键导入表情：

非docker：
RAILS_ENV=production bin/tootctl emoji import --category=自定义表情分类名 表情tar.gz文件地址

docker：
docker cp ./表情地址 web服务docker名:/tmp/表情名.tar.gz
docker-compose exec web bin/tootctl emoji import --category=自定义表情分类名 /tmp/表情名.tar.gz`)
	Exit()
}

func Exit() {
	starainrt.MsgBox("按回车键退出", "")
	os.Exit(0)
}

func Chose01(proxy string) error {
	var url string
	for {
		url = starainrt.MsgBox("\n请输入实例域名（如:mastodon.social）", "")
		if url == "" {
			fmt.Println("输入不合法，请重新输入")
			continue
		}
		break
	}
	fmt.Println("获取并解析表情中，请稍后………")
	data, err := GetEmojiNormal("https://"+url, proxy)
	if err != nil {
		return err
	}
	fetchEmojiLists, err = AnalyseEmojis(data)
	fmt.Printf("\n下载并解析完成，共获取到%d个表情\n\n", len(fetchEmojiLists))
	return err
}

func Chose02() error {
	var err error
	var url string
	for {
		url = starainrt.MsgBox("请输入实例域名（如:mastodon.social）", "")
		if url == "" {
			fmt.Println("输入不合法，请重新输入")
			continue
		}
		break
	}
	fmt.Println("请在浏览器登录此实例后，访问下面的URL，复制里面所有的内容并保存到本地txt中")
	fmt.Println("https://" + url + "/api/v1/custom_emojis")
	pasteStr := starainrt.MessageBox("请输入保存在本地的文件地址（可拖拽到这里）：", "")
	data, err := GetEmojiFromFile(pasteStr)
	if err != nil {
		return err
	}
	fetchEmojiLists, err = AnalyseEmojis(data)
	fmt.Printf("\n解析完成，共获取到%d个表情\n", len(fetchEmojiLists))
	return err
}

func Usagi() {
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

# bili-music-player

---

### 项目描述：该项目是一个前端**vue+Wails**，后端**go**的软件；对于**vue+Wails+go**而言我是零基础；我想通过后端获取bilibili网页的json格式的信息，提取出一个视频合集的所有bv号以及标题、音频链接，然后由一套算法得出歌曲名称以及歌手，再用歌曲名称以及歌手信息通过我指定的api获取歌词，然后由前端列表把指定的歌曲作者、歌曲名称、歌词、音频链接显示出来，制作一个简单的网页音乐播放器。

---

## 当前进度：正在编写前端

---



## 后端：

### 已完成：

1. 类：**bili**

2. 函数**bili_init**, 存入**cookie**等

3. 函数**bili_try**, 尝试访问 **https://www.bilibili.com/**, 访问成功返回**true**, 失败返回**false**

4. 函数**bili_get_pi**, 接收一个文本参数如 **BV1oU1jBXEN8** 和另一个文本参数如 **pi.json**, 然后访问 **https://www.bilibili.com/video/BV1oU1jBXEN8** , 获取其 **response**, 然后用正则匹配**...**处的内容：
   
   ```
       <script>
          window.__playinfo__ = {
              **...**
           }
       </script>
   ```
   
   获取到后返回文本数据，写入本地文件**pi.json**中

5. 函数**bili_get_is**, 与**bili_get_pi**相似, 只不过把 ```window.__playinfo__``` 换成 ```window.__INITIAL_STATE__```

6. 函数**bili_js**, 接收文本参数 如 **pi.json**, 读入**pi.json**然后将其**json**格式化, 获取其**json**格式化的信息以便后续调用

7. 写一个函数func GetNestedString(data map[string]any, keys ...string) (string, bool), 用于获取json数据的元素信息

8. GetUGCSeasonTitle(defaultISPath) 输出标题：

   - 优先尝试 ParseJSON + GetNestedString(jsInfo, "ugc_season", "title")
   - 如果 is.json 不是严格 JSON，就回退到原文正则提取 ugc_season.title
   - 没有则输出空

9. 函数**bili_get_bmpinfo**
    - 读入**bili_get_is**返回的**json数据: is.json**, 然后解析数据, 获取数据中的(对应python格式) **is.json['ugc_season'][0]['episodes']** (数组)
    - 注意若未读取到 **ugc_season** 的时候, 返回
    - 对**episodes**遍历，获取**episodes[i]**中的title, pic, bvid, audio(暂时留空, 供**bili_bmpinfo_fix**使用)
    - 记录 **ugcTitle = GetUGCSeasonTitle**
    - 若**bmpinfo.json**不存在, 则创建新的我自己的json数据：```{ugcTitle{"title":..., "pic": ..., ......}]}```(当中的ugcTitle和bvid是变量名) 然后输出到**bmpinfo.json**中
    - 若**bmpinfo.json**存在, 且**bmpinfo.json**不存在**ugcTitle**, 则向**bmpinfo.json**添加新的ugcTitle项, 否则返回即可

10. 函数**bili_get_audio**, 读pi.json数据, 先得到```pi['data']['dash']['audio']```数组**audio_info**, 对其进行遍历, 返回**audio_info['bandwidth']**(整数型)值最大的那一项的audio_info['baseUrl']

11. 函数**bili_audio_download**, 接收参数**url**和**file_path**, 下载url的二进制文件存储到file_path, 下载时记得带
```"origin": "https://www.bilibili.com"``` 否则会被拒绝访问, 访问成功则开始下载

12. 函数**bili_bmpinfo_fix**, 接收参数**bv**, 读取**bmpinfo.json**
    - bv号不为空, 对其中的每个标题的每个bvid数据进行遍历查找, 若匹配，则bili_get_pi(bv), url = bili_get_audio, 写入**bmpinfo.json**, 把对应的留空的audio填上"audio":url
    - 若bv号为空, 则读入**bmpinfo.json**, 遍历每个bvid数据（遍历延迟2秒），然后将所有的bv号都填上对应的"audio":url
    - 注意若已存在"audio":url, 则判断**audio**链接是否超出时效性, **audio**如下：
        "audio": "https://cn-sccd-fx-01-03.bilivideo.com/upgcxcode/57/39/1360673957/1360673957-1-30232.m4s?e=ig8euxZM2rNcNbdlhoNvNC8BqJIzNbfqXBvEqxTEto8BTrNvN0GvT90W5JZMkX_YN0MvXg8gNEV4NC8xNEV4N03eN0B5tZlqNxTEto8BTrNvNeZVuJ10Kj_g2UB02J0mN0B5tZlqNCNEto8BTrNvNC7MTX502C8f2jmMQJ6mqF2fka1mqx6gqj0eN0B599M=\u0026uipk=5\u0026oi=3737470867\u0026og=hw\u0026nbs=1\u0026platform=pc\u0026gen=playurlv3\u0026os=bcache\u0026trid=0000dacce7a9f4c94d41bda6009deea4f30u\u0026mid=0\u0026deadline=1776602022\u0026upsig=688e265e1c4b8cf4a80f17d500526974\u0026uparams=e,uipk,oi,og,nbs,platform,gen,os,trid,mid,deadline\u0026cdnid=3903\u0026bvc=vod\u0026nettype=0\u0026bw=107980\u0026f=u_0_0\u0026qn_dyeid=39c44577e2fbc67b003e983369e4af86\u0026agrr=1\u0026buvid=3B251CCF-63C2-12DC-1F6C-6DC3DFF5689322457infoc\u0026np=151404637\u0026build=0\u0026dl=0\u0026orderid=0,3"
        判断是否超出时效性，如没超出则返回，若超出则更新**audio**, 重新填入

13. 函数**bili_lget_is** 和 **bili_lget_pi** 
    - 与**bili_get_is**和**bili_get_pi**相似, 只不过参数变成了**ml3888553754**
    - 访问地址变成了**https://www.bilibili.com/list/ml3888553754**, response中同样有```window.__playinfo__``` 和 ```window.__INITIAL_STATE__```
    - 与**GetUGCSeasonTitle**类似, 写一个函数**GetListTitle** 获取**is.json**中的```['mediaListInfo']['title']```, 失败返回空

14. 函数**bili_lget_bmpinfo**
    - 读入**is.json**，然后解析数据，获取数据中的(对应python格式) **is.json['resourceList']** (数组)
    - 对**resourceList**遍历，获取**resourceList[i]**中的title, cover, bvid, audio(暂时留空, 供**bili_bmpinfo_fix**使用)
    - 记录 ```listTitle = GetListTitle```
    - 若**bmpinfo.json**不存在, 则创建新的我自己的json数据：```{listTitle[{"title":..., "pic": ..., ......}]}```(当中的listTitle和listid是变量名) 然后输出到**bmpinfo.json**中
    - 这里的**pic**就是获取到的**cover**, 但是要在前面手动加上**http:**, 因为**cover**是**//**开头的
    - 若**bmpinfo.json**存在, 且**bmpinfo.json**不存在**listTitle**, 则向**bmpinfo.json**添加新的listTitle项, 否则返回即可

15. 函数**bili_del**, 接收参数 **title**, 读入**bmpinfo.json**, 删除**bmpinfo.json**中的**title**元素(title指某个歌单的标题)
---

### 待完成：



完成之后生成命令行文件 **bilig.exe**

### 命令行：
静默执行：**bili_init** 
1. bilig -try
    - 执行**bili_try**
    - 输出(可选)**[true, false]**
2. bilig -get [bv]
    - 执行**bili_get_pi(bv)**
    - 执行**bili_get_is(bv)**
    - 执行**bili_get_bmpinfo**
4. bilig --title
    - 输出 **GetUGCSeasonTitle**
5. bilig -fix [bv]
    - 执行**bili_bmpinfo_fix(bv)**
5. bilig -fix--all
    - 执行**bili_bmpinfo_fix(bv留空)** 
6. bilig -download [url] [title]
    - 执行**bili_audio_download(url, title)** 
7. bilig -lget [ml]
    - 执行**bili_lget_is(ml)**
    - 执行**bili_lget_bmpinfo**

---
### 修正项：

---

### 注意事项：

1. 代码不要写在单个文件中，注意分好类，提高代码复用性与可阅读性
2. 有什么不好判断的信息可以先问我
3. 我的仓库已经创建好：**git@github.com:huginmost/bili-music-player.git**，可以建立分支测试，我确定可用后再让你合并
4. 可以设置全局变量控制is.json或pl.json的路径名，我写出来只是为了方便理解，但其实不用加到函数(如**GetUGCSeasonTitle**、**bili_get_bmpinfo**等)的参数中去，真实调用时直接就是GetUGCSeasonTitle() bili_get_bmpinfo()等


## 前端
vue + vite
1. 搜索网页上的一些本地音乐播放器模板, 根据模板制作一个简单模版ui, 我想要一个偏夜间风格的
2. 阅读**bmpinfo.json**, 里面存储了歌单标题, 歌曲标题, 封面等信息, 根据该json数据进行制作
3. 利用后端中的 **bilig -fix [bv]** 可以获取audio链接, 然后读入bmpinfo.json中的audio对音乐进行播放
4. 后端中已完成的函数在本文档中
5. 要求该网页音乐播放器有随机播放，顺序播放功能，以及其它基础功能（参考网页上搜的模版）
6. 对歌曲进行预缓存，可以用队列，缓存数量在3左右，也就是除了正在播放的歌曲用-fix，播放列表中的下3首都用-fix获取了链接，保障播放下一首时能"秒反应"
7. 有删除歌单功能（对bmpinfo中的某个歌单元素进行删除）和下载功能（下载某个audio）
8. 可以多制作几个版本，我将选取最优的版本发布并使用
9. 注意前端访问时也要带**origin**, 否则会拒绝访问
10. 可以调整音量

将编译出来的的文件打包进**release**文件夹中
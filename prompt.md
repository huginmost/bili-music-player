# bili-music-player

---

### 项目描述：该项目是一个前端**vue+Wails**，后端**go**的软件；对于**vue+Wails+go**而言我是零基础；我想通过后端获取bilibili网页的json格式的信息，提取出一个视频合集的所有bv号以及标题、音频链接，然后由一套算法得出歌曲名称以及歌手，再用歌曲名称以及歌手信息通过我指定的api获取歌词，然后由前端列表把指定的歌曲作者、歌曲名称、歌词、音频链接显示出来。

---

### 当前进度：正在编写后端

---

### 已完成：

1. 类：**bili**

2. 函数**bili_init**, 存入**cookie**

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

7. 写一个函数func getNestedString(data map[string]any, keys ...string) (string, bool), 用于获取json数据的元素信息

8. GetUGCSeasonTitle(defaultISPath) 输出标题：
   
   - 优先尝试 ParseJSON + GetNestedString(jsInfo, "ugc_season", "title")
   - 如果 is.json 不是严格 JSON，就回退到原文正则提取 ugc_season.title

9. 函数**bili_get_bmpinfo**, 获取**bili_get_is**返回的**json数据**，然后解析数据，获取数据中的(对应python格式) **['ugc_season'][0]['episodes']** (数组)，然后对**episodes**遍历，获取**episodes[i]**中的title, pic, bvid; 创建新的我自己的json数据：```{plinfo[{"title":..., "pic": ..., ......}]}``` 然后输出到bmpinfo.json中

10. 函数**bili_get_audio**, 读pi.json数据, 先得到```pi['data']['dash']['audio']```数组**audio_info**, 对其进行遍历, 返回**audio_info['bandwidth']**(整数型)值最大的那一项的audio_info['baseUrl']

11. 函数**bili_audio_download**, 接收参数**url**和**file_path**, 下载url的二进制文件存储到file_path, 下载时记得带
```"origin": "https://www.bilibili.com"``` 否则会被拒绝访问, 访问成功则开始下载
---

### 待完成：


完成之后做**main**任务

#### main：

写函数**main**：
1. **bili_init**
2. 输出**bili_try**的返回值，若为**false**则返回
3. **bili_get_pi(BV1oU1jBXEN8, pi.json)**, **bili_get_is(BV1oU1jBXEN8, is.json)**
4. 输出**GetUGCSeasonTitle(is.json)**的内容
5. **bili_get_bmpinfo(is.json)** 导出 **bmpinfo.json**
6. url = bili_get_audio(pi.json), 输出url, bili_audio_download(url, test.m4a)
7. 结束

2. 生成可执行文件 **bilig.exe**

---

### 修正项：

---

### 注意事项：

1. 代码不要写在单个文件中，注意分好类，提高代码复用性与可阅读性
2. 有什么不好判断的信息可以先问我
3. 我的仓库已经创建好：**git@github.com:huginmost/bili-music-player.git**，可以建立分支测试，我确定可用后再让你合并
4. 可以设置全局变量控制is.json或pl.json的路径名，我写出来只是为了方便理解，但其实不用加到函数(如**GetUGCSeasonTitle**、**bili_get_bmpinfo**等)的参数中去，真实调用时直接就是GetUGCSeasonTitle() bili_get_bmpinfo()等

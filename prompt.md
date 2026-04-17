# bili-music-player
---
### 项目描述：该项目是一个前端**vue+Wails**，后端**go**的软件；对于**vue+Wails+go**而言我是零基础；我想通过后端获取bilibili网页的json格式的信息，提取出一个视频合集的所有bv号以及标题、音频链接，然后由一套算法得出歌曲名称以及歌手，再用歌曲名称以及歌手信息通过我指定的api获取歌词，然后由前端列表把指定的歌曲作者、歌曲名称、歌词、音频链接显示出来。
---
### 当前进度：正在编写后端
---
### 已完成：
1. 创建类：**bili**

2. 写一个函数**bili_init**, 存入**cookie**

3. 写一个函数**bili_try**, 尝试访问 **https://www.bilibili.com/**, 访问成功返回**true**, 失败返回**false**

4. 写一个函数**bili_get_pi**, 接收一个文本参数如 **BV1oU1jBXEN8** 和另一个文本参数如 **pi.json**, 然后访问 **https://www.bilibili.com/video/BV1oU1jBXEN8**, 获取其 **response**, 然后用正则匹配**...**处的内容：

    ```
        <script>
           window.__playinfo__ = {
               **...**
            }
        </script>
    ```
获取到后返回文本数据，写入本地文件**pi.json**中

5. 写一个函数**bili_js**, 接收文本参数 如 **pi.json**, 读入**pi.json**然后将其**json**格式化, 获取其**json**格式化的信息以便后续调用

6. 写一个函数**bili_get_is**, 与**bili_get_pi**相似, 只不过把 ```window.__playinfo__``` 换成 ```window.__INITIAL_STATE__```


---
### 待完成：
1. 写一个函数func getNestedString(data map[string]any, keys ...string) (string, bool)

    
完成之后做**main**任务

#### main：

1. 写函数**main**, 先**bili_init**, 然后输出**bili_try**的返回值，若为**true**则**bili_get_pi(BV1oU1jBXEN8, pi.json)**, **bili_get_is(BV1oU1jBXEN8, is.json)**, 然后添加
```
    title, ok := getNestedString(jsInfo, "channelKv", "ugc_season", "title")
    if !ok {
    	log.Fatal("title not found")
    }
    fmt.Println(title)
```
, 结束
   
2. 生成可执行文件 **bilig.exe**
---
### 修正项：
无
---
### 注意事项：

1. 代码不要写在单个文件中，注意分好类，提高代码复用性与可阅读性
2. 有什么不好判断的信息可以先问我
3. 我的仓库已经创建好：**git@github.com:huginmost/bili-music-player.git**，可以建立分支测试，我确定可用后再让你合并

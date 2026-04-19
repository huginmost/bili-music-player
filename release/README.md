# Release

本目录用于存放可直接分发的构建产物。

约定内容：

- `bilig.exe`
- `bmplayer-web.exe`
- `frontend/` 静态资源

重新打包时可执行：

```powershell
go build -o bilig.exe ./cmd/bilig
go build -o bmplayer-web.exe ./cmd/bmplayer-web
cd frontend
npm run build
```

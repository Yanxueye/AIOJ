# remote_judge Docker Images

本目录用于构建 `remote_judge` 本地评测镜像。

## 镜像列表

- `remote-judge-cpp17`
- `remote-judge-go122`
- `remote-judge-python311`

## 构建方式

```powershell
Set-Location C:\Users\17354\Desktop\项目实训\remote_judge
.\docker\build-images.ps1
```

## 说明

构建完成后，`remote_judge` 将优先使用这些本地镜像进行真实代码评测，而不是依赖运行时临时拉取公共镜像。

# nasa-apod-cli

一条命令获取 NASA 每日天文图片（[APOD](https://apod.nasa.gov/)）的链接。

Go 单文件实现，无第三方依赖。数据源为官方 [science.nasa.gov APOD 接口](https://science.nasa.gov/apod)（APOD 于 2026-11 从 apod.nasa.gov 迁移而来），无需 API key。

**主页**：<https://pidgey016.github.io/nasa-apod-cli/> · **下载**：[Releases](../../releases/latest)

## 使用

```bash
nasa-apod                    # 输出当天图片的标题和链接
nasa-apod --date 2026-11-19  # 指定日期
nasa-apod --open             # 同时在浏览器中打开
nasa-apod --version
```

默认日期为美东时间（APOD 发布时区）的当天。当天图片尚未发布时接口返回 404 并报错退出，可用 `--date` 指定已发布的日期。

输出示例：

```
NASA Astronomy Picture of the Day (2026-11-20)
Title: The Butterfly Nebula from Hubble
Link:  https://apod.nasa.gov/apod/image/2611/butterfly_hubble_1920.jpg
```

## 安装

从 [Releases](../../releases/latest) 下载对应平台的压缩包：

| 平台 | 架构 | 产物 |
|---|---|---|
| macOS | arm64 / amd64 | `.tar.gz` |
| Linux | amd64 / arm64 | `.tar.gz` |
| Windows | amd64 | `.zip` |

```bash
# macOS / Linux
tar xzf nasa-apod_v0.1.0_darwin_arm64.tar.gz
sudo mv nasa-apod /usr/local/bin/
```

## 从源码构建

```bash
go build -o nasa-apod .
```

## 发布

推送 `v*` tag 即可，GitHub Actions 自动交叉编译 5 个平台产物并发布 Release：

```bash
git tag v0.1.0 && git push origin v0.1.0
```

## License

MIT
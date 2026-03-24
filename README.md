# bcfntconv

将 Nintendo 3DS 的 `bcfnt` 与模拟器使用的 `shared_font.bin` 相互转换，并提供将 `bcfnt` 打包为 3DS 可安装 `.cia` 的工具 `bcfnt2cia`（需要外部 `3dstool` 与 `make_cia`）。

## 快速构建

```bash
go build -trimpath -ldflags "-s -w" ./cmd/bcfnt2bin -o bin/bcfnt2bin.exe
go build -trimpath -ldflags "-s -w" ./cmd/bin2bcfnt -o bin/bin2bcfnt.exe
go build -trimpath -ldflags "-s -w" ./cmd/bcfnt2cia -o bin/bcfnt2cia.exe
```

已包含的预编译 Windows 二进制位于 `bin/`。

## `bcfnt` 与 `shared_font.bin` 相互转换

用法示例：

- 将 `bcfnt` 转为 `shared_font.bin`：

```
./bin/bcfnt2bin.exe input.bcfnt output_shared_font.bin
```

- 将 `shared_font.bin` 转回 `bcfnt`：

```
./bin/bin2bcfnt.exe input_shared_font.bin output.bcfnt
```

实现说明（简要）：
- 将 `bcfnt` 的第 4 个字节从 `0x54` 改为 `0x55`；
- 在前面添加 0x80 字节的头部：头部第 1 字节为 `0x02`，第 5 字节为 `0x01`，第 9-12 字节写入原始 `bcfnt` 长度（little-endian），其它为 0；
- 反向转换则去掉前 0x80 字节头部，并把第 4 个字节改回 `0x54`。

## `bcfnt` 打包成 3DS `.cia`

```
./bin/bcfnt2cia.exe input.bcfnt [output.cia]
```

依赖：
- `3dstool`、`make_cia`（请确保这两个工具在 `PATH` 中，或将它们放置在与 `bcfnt2cia.exe` 相同的目录，程序会优先使用同目录的工具）。

输入校验（重要）：
- 输入 `bcfnt` 必须以 ASCII `CFNT` 开头（程序会校验前 4 字节）。
- 为避免生成不合法或过大的中间文件，程序会拒绝长度超过 `0x332000 - 0x80` 的 `bcfnt` 文件（会报错并中止）。

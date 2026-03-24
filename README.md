# bcfntconv

工具：将 Nintendo 3DS 的 `bcfnt` 与模拟器使用的 `shared_font.bin` 相互转换。

构建：

```bash
go build ./cmd/bcfnt2bin -o bcfnt2bin
go build ./cmd/bin2bcfnt -o bin2bcfnt
```

Windows 可交叉编译示例：

```bash
env GOOS=windows GOARCH=amd64 go build ./cmd/bcfnt2bin -o bcfnt2bin.exe
env GOOS=windows GOARCH=amd64 go build ./cmd/bin2bcfnt -o bin2bcfnt.exe
```

用法：

- 将 `bcfnt` 转为 `shared_font.bin`：

```
./bcfnt2bin input.bcfnt output_shared_font.bin
```

- 将 `shared_font.bin` 转回 `bcfnt`：

```
./bin2bcfnt input_shared_font.bin output.bcfnt
```

实现说明：

- 将 `bcfnt` 的第 4 个字节（1-based）从 `0x54` 改为 `0x55`；
- 在前面添加 0x80 字节的头部：头部第 1 字节为 `0x02`，第 5 字节为 `0x01`，第 9-12 字节写入原始 `bcfnt` 长度（little-endian），其它为 0；
- 反向转换则去掉前 0x80 字节头部，并把第 4 个字节改回 `0x54`。
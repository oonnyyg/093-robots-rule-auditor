# robots.txt 离线审查工具 — 输入/输出格式

## 运行

```sh
go run ./cmd/robots input.json
```

仅使用 Go 标准库。结果以 JSON 写入标准输出；输入非法时向标准错误
输出错误信息并以非零状态退出，不输出结果。

## 输入格式

```json
{
  "robots": "User-agent: Googlebot\nDisallow: /private\n",
  "queries": [
    {"crawler": "Googlebot", "path": "/private/x?a=1"}
  ]
}
```

- `robots`：robots.txt 规则文本（字符串，可含换行）。
- `queries`：查询数组，按原顺序处理。
  - `crawler`：爬虫名，非空，仅含 ASCII 字母、`_`、`-`。
  - `path`：候选路径，以 `/` 开头，仅含可见 ASCII（无空白、无 `%`），
    可带查询串（`?...`），不含片段（`#`）。

## 输出格式

```json
{
  "results": [
    {
      "crawler": "Googlebot",
      "path": "/private/x?a=1",
      "allowed": false,
      "ruleLine": 2,
      "groups": [1]
    }
  ]
}
```

- `allowed`：是否允许抓取。
- `ruleLine`：决定结论的规则原始行号（1 起始）；无选中规则时为 `null`。
- `groups`：参与裁决的组起始行号（即各组首个 `User-agent` 行号）。

## 解析规则

- 每行先去除 `#` 起的行内注释，再去除首尾空白；空行不参与分组。
- 指令名忽略大小写；未知指令与出现在任何组之外的规则被忽略。
- `User-agent` 值为 `*` 或合法爬虫名；连续的 `User-agent` 行属于同组，
  出现 `Allow`/`Disallow` 后再遇 `User-agent` 另起一组。
- 空规则不参与匹配；非空规则必须以 `/` 开头，否则视为格式非法。
- 任何已知项（爬虫名、路径、User-agent 值、规则模式）格式非法都会
  导致非零退出。

## 匹配与裁决

- 按爬虫名忽略大小写完整匹配并合并所有命中组；无命中时合并所有
  `User-agent: *` 的组；仍无可用组时默认允许。
- 路径区分大小写；规则从路径首字符开始匹配；`*` 匹配任意长度字符；
  末尾 `$` 要求匹配到路径结尾，其他位置不允许出现 `$`；无 `$` 时
  允许剩余后缀。
- 命中规则按“除 `*` 和末尾 `$` 之外的字符数”取最长；平局时 `Allow`
  优先；仍同类时取原始行号最小者。
- 无命中规则默认允许；`/robots.txt` 始终允许。

## 示例

见仓库根目录 `sample.json`：

```sh
go run ./cmd/robots sample.json
```

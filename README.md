# robots-rule-auditor

离线 robots.txt 审查工具：读取规则文本与爬虫路径查询，逐条报告是否允许抓取及判定依据。仅使用 Go 标准库。

## 运行

```
go run ./cmd/robots 输入.json
```

结果以 JSON 写到标准输出。输入格式非法（含查询项非法）时向标准错误输出错误并以非零码退出，标准输出不产生任何内容。

## 输入格式

```json
{
  "rules": "User-agent: bot\nDisallow: /admin\n",
  "queries": [
    {"crawler": "bot", "path": "/admin/panel"},
    {"crawler": "other_bot", "path": "/news?p=1"}
  ]
}
```

- `rules`：robots.txt 全文，字符串。
- `queries[].crawler`：爬虫名，非空，仅含 ASCII 字母、`_`、`-`。
- `queries[].path`：以 `/` 开头的候选路径，仅含可见 ASCII 字符（无空白、无 `%`），可带查询串，不含 `#` 片段。

## 输出格式

```json
{
  "results": [
    {"allowed": false, "ruleLine": 2, "groups": [1]},
    {"allowed": true, "ruleLine": null, "groups": [4, 7]}
  ]
}
```

- `results` 与查询一一对应，保持原顺序。
- `allowed`：是否允许抓取。
- `ruleLine`：决定结论的规则在 `rules` 中的 1 起始行号；无命中规则时为 `null`。
- `groups`：本次参与判定的组起始行号（组首条 `User-agent` 所在行），按出现顺序排列。

## 判定规则

- 指令名忽略大小写；去除行内 `#` 注释与首尾空白；未知指令、组外规则、非法 `User-agent` 值（非 `*` 且非合法爬虫名）均忽略。
- 连续 `User-agent` 行属于同组；出现 `Allow`/`Disallow` 行后再遇 `User-agent` 另起新组；空行不分组。
- 按爬虫名忽略大小写完整匹配并合并所有命中组；无命中时合并所有 `*` 组；仍无可用组则默认允许。
- 空规则与非 `/` 开头的规则不参与匹配；`$` 仅允许出现在规则末尾（表示匹配到路径结尾），其余位置出现则该规则无效。
- 路径区分大小写，规则从路径首字符开始匹配，`*` 匹配任意长度字符；无 `$` 时允许剩余后缀。
- 命中规则中按“除 `*` 和末尾 `$` 外的字符数”取最长；平局 `Allow` 优先；仍平局取原始行号最小者。无命中默认允许；`/robots.txt` 始终允许。

## 测试

```
go test ./...
```

# Contributing to GEO Optimizer

感谢你考虑为 GEO Optimizer 做出贡献！

## 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发环境设置](#开发环境设置)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)

## 行为准则

- 尊重所有贡献者
- 保持专业、建设性的讨论
- 接受建设性批评

## 如何贡献

### 报告 Bug

如果你发现了 bug，请通过 [GitHub Issues](https://github.com/geo-team-red/geo-optimizer/issues) 提交，包含：

1. **问题描述** - 清晰描述发生了什么
2. **复现步骤** - 如何触发这个 bug
3. **期望行为** - 你期望发生什么
4. **实际行为** - 实际发生了什么
5. **环境信息** - Go 版本、操作系统等

### 提出新功能

新功能建议请通过 Issues 提交，描述：

1. **使用场景** - 这个功能解决什么问题
2. **建议方案** - 你认为应该如何实现
3. **替代方案** - 是否有其他解决方式

### 提交代码

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 开发环境设置

### 前置要求

- Go 1.21+
- Git
- LLM API Key（用于集成测试）

### 克隆与构建

```bash
# 克隆你的 fork
git clone https://github.com/YOUR_USERNAME/geo-optimizer.git
cd geo-optimizer

# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建
go build ./...
```

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包测试
go test ./pkg/optimizer -v

# 运行单个测试
go test ./pkg/analyzer -run TestScorer -v
```

### 代码格式化

```bash
# 格式化代码（提交前必须执行）
go fmt ./...
```

## 代码规范

### 项目结构

```
pkg/
├── optimizer/         # 核心优化引擎
│   └── strategies/    # 策略实现
├── llm/               # LLM 客户端抽象
│   └── prompts/       # Prompt 模板
├── models/            # 数据模型
├── analyzer/          # 内容评分
└── config/            # 配置与预设
```

### 添加新策略

实现 `Strategy` 接口：

```go
type Strategy interface {
    Name() string
    Type() models.StrategyType
    Validate(req *models.OptimizationRequest) bool
    Preprocess(content string, req *models.OptimizationRequest) string
    Postprocess(content string, req *models.OptimizationRequest) string
    BuildPrompt(req *models.OptimizationRequest) string
}
```

示例：

```go
package strategies

import (
    "github.com/geo-team-red/geo-optimizer/pkg/models"
    "github.com/geo-team-red/geo-optimizer/pkg/llm/prompts"
)

type MyStrategy struct {
    *BaseStrategy
}

func NewMyStrategy() *MyStrategy {
    return &MyStrategy{
        BaseStrategy: NewBaseStrategy("my_strategy", "My Strategy"),
    }
}

func (s *MyStrategy) Validate(req *models.OptimizationRequest) bool {
    return req.Content != ""
}

func (s *MyStrategy) BuildPrompt(req *models.OptimizationRequest) string {
    return prompts.NewBuilder().BuildStrategyPrompt(models.StrategyType("my_strategy"), req)
}
```

### 添加新 LLM Provider

实现 `LLMClient` 接口：

```go
type LLMClient interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
```

### 代码风格

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 为导出的函数和类型添加注释
- 错误信息不要大写开头，不要以标点结尾

## 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Type 类型

| Type | 描述 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响功能） |
| `refactor` | 代码重构 |
| `test` | 测试相关 |
| `chore` | 构建、依赖等 |

### 示例

```
feat(optimizer): add keyword density strategy

- Implement KeywordDensityStrategy
- Add configuration for density threshold
- Add unit tests

Closes #123
```

## Pull Request 流程

1. **确保测试通过** - `go test ./...`
2. **格式化代码** - `go fmt ./...`
3. **更新文档** - 如有 API 变更，更新 `API.md`
4. **添加测试** - 新功能需要测试覆盖
5. **一个 PR 一个功能** - 保持 PR 聚焦

### PR 标题格式

```
<type>(<scope>): <description>
```

示例：`feat(strategies): add keyword optimization strategy`

### Review 检查项

- [ ] 代码通过所有测试
- [ ] 代码已格式化
- [ ] 新代码有适当注释
- [ ] 文档已更新
- [ ] 没有引入新的警告

## 许可证

通过提交代码，你同意你的贡献将根据 [MIT License](LICENSE) 授权。

---

有问题？随时在 [Issues](https://github.com/geo-team-red/geo-optimizer/issues) 中提问！

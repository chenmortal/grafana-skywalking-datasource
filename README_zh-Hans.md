# Skywalking Grafana 数据源插件

![Dynamic JSON Badge](https://img.shields.io/badge/dynamic/json?logo=grafana&query=%24.version&url=https%3A%2F%2Fgrafana.com%2Fapi%2Fplugins%2Fchenmortal-skywalking-datasource&label=Marketplace&prefix=v&color=F47A20)
![Dynamic JSON Badge](https://img.shields.io/badge/dynamic/json?logo=grafana&query=%24.grafanaDependency&url=https%3A%2F%2Fgrafana.com%2Fapi%2Fplugins%2Fchenmortal-skywalking-datasource&label=Grafana&color=F47A20)

[English](https://github.com/chenmortal/grafana-skywalking-datasource/blob/master/README.md) | 简体中文

Grafana 数据源插件，通过 GraphQL API 连接 **Apache SkyWalking** OAP 服务器，实现在 Grafana 中直接查看分布式链路追踪数据。

## 概述

本插件将 SkyWalking 的分布式追踪数据引入 Grafana，支持以下功能：

- **搜索链路** — 按 Layer、Service、Endpoint、Instance 等维度搜索
- **查看链路详情** — 使用 Grafana 原生 Trace 视图展示完整调用链
- **链路状态筛选** — 按成功/失败状态过滤，按耗时或开始时间排序
- **搜索结果跳转** — 通过内置数据链接从搜索结果直接跳转到详细链路视图

插件同时支持 SkyWalking v1（`queryBasicTraces` / `queryTrace`）和 v2（`queryTraces`）GraphQL API，可通过开关切换。

## 环境要求

- **Grafana** `>= 12.1.0`
- **Apache SkyWalking OAP** 服务器，Grafana 可访问其 GraphQL 端点

## 快速开始

### 安装

1. 从 [Grafana 插件市场](https://grafana.com/grafana/plugins/chenmortal-skywalking-datasource/) 安装，或下载发布包。
2. 将插件放入 Grafana 的 plugins 目录（参考 [Grafana 插件安装指南](https://grafana.com/docs/grafana/latest/administration/plugin-management/)）。

### 配置数据源

1. 在 Grafana 中进入 **Administration** → **Plugins and data** → **Data sources** → **Add data source**。
2. 搜索 **Skywalking** 并选中。
3. 填写配置：

| 配置项                   | 说明                                | 示例                                        |
| ------------------------ | ----------------------------------- | ------------------------------------------- |
| **URL**                  | SkyWalking OAP GraphQL 端点地址     | `http://skywalking-oap.example.com/graphql` |
| **Path**                 | 附加资源路径                        | `/resources`（默认值）                      |
| **Interface Version v2** | 启用 SkyWalking v2 Query Traces API | `true` / `false`                            |
| **API Key**              | 认证令牌（可选）                    | `your-api-key`                              |

> **注意：** 开启 Interface Version v2 后无法回退。此设置决定使用哪个版本的 SkyWalking API 进行链路查询。

### 通过 Provisioning 配置

```yaml
datasources:
  - name: 'skywalking'
    type: 'chenmortal-skywalking-datasource'
    access: proxy
    url: 'http://skywalking-oap:12800/graphql'
    jsonData:
      path: '/resources'
      interfacev2: false
    secureJsonData:
      apiKey: 'your-api-key'
```

## 查询类型

插件支持两种查询模式：

### 搜索模式

按以下维度搜索链路：

- **Layer** — 选择层级（如 GENERAL、VIRTUAL_MQ 等）
- **Service** — 按服务名筛选
- **Endpoint** — 按关键字搜索端点
- **Instance** — 按服务实例筛选
- **Trace State** — 全部 / 成功 / 失败
- **Query Order** — 按耗时排序（最慢优先）或按开始时间排序（最新优先）

### TraceID 查询

输入 Trace ID 查找特定链路。链路将使用 Grafana 原生 Trace 视图展示，包含完整的 Span 详情：服务标签、Span 标签、引用关系和耗时信息。

## 仪表盘集成

链路搜索结果包含指向详细链路视图的链接。选中某条链路后，Grafana 会打开详细的 Trace 视图，展示：

- 包含父子关系的 Span 树
- 操作名称、服务名称和实例信息
- 每个 Span 的耗时分解
- Span 标签（peer、layer、component 及自定义标签）
- 引用关系（CHILD_OF、FOLLOWS_FROM）

## 开发指南

### 环境准备

- **Go** `>= 1.26`
- **Node.js** `>= 22`
- **npm** `>= 11`
- **Mage**（Go 构建工具）

### 初始化

```bash
# 安装前端依赖
npm install

# 安装 Go 依赖
go mod download

# 生成后端 GraphQL 类型
go generate ./pkg/...

# 生成前端 GraphQL 类型
npm run codegen
```

### 构建

```bash
# 构建前后端
npm run build

# 开发模式（热更新）
npm run dev
```

### 测试

```bash
# 前端单元测试
npm test

# 前端 CI 测试
npm run test:ci

# Go 后端测试
go test ./...

# 端到端测试
npm run e2e
```

### 代码检查

```bash
# TypeScript 类型检查
npm run typecheck

# ESLint
npm run lint

# ESLint 自动修复
npm run lint:fix
```

### Docker 本地开发

```bash
# 启动 Grafana 并挂载插件
npm run server
```

此命令通过 Docker Compose 启动一个已加载本插件的 Grafana 实例。

### 项目结构

```
├── pkg/                          # Go 后端插件
│   ├── main.go                   # 入口文件
│   └── plugin/
│       ├── datasource.go         # 数据源生命周期
│       ├── skywalking.go         # 服务层、健康检查
│       ├── client.go             # GraphQL API 客户端
│       ├── query.go              # 查询处理与链路帧转换
│       ├── callresource.go       # HTTP 资源路由
│       └── graphql.go            # 生成的 GraphQL 类型
├── src/                          # 前端 TypeScript/React
│   ├── module.ts                 # 插件注册
│   ├── datasource.ts             # 数据源类
│   ├── components/
│   │   ├── ConfigEditor.tsx      # 数据源配置界面
│   │   ├── QueryEditor.tsx       # 查询构建器
│   │   └── SearchForm.tsx        # 链路搜索表单
│   └── locales/                  # 国际化翻译（en-US, zh-Hans）
├── graphql/                      # GraphQL Schema 与查询定义
├── provisioning/                 # Grafana Provisioning 示例
└── tests/                        # Playwright 端到端测试
```

后端使用 Go 编写，通过 GraphQL 与 SkyWalking OAP 服务器通信。前端是 TypeScript/React 应用，提供查询编辑器 UI 并与 Grafana 的链路可视化集成。

### GraphQL 代码生成

项目使用代码生成工具管理前后端的 GraphQL 类型：

- **后端：** [genqlient](https://github.com/Khan/genqlient) 根据 `graphql/` 中的 GraphQL 操作生成 Go 类型
- **前端：** [@graphql-codegen](https://the-guild.dev/graphql/codegen) 根据相同的 GraphQL Schema 生成 TypeScript 类型

Schema 变更后重新生成类型：

```bash
# 后端
go generate ./pkg/...

# 前端
npm run codegen
```

## 参与贡献

欢迎提交 Issue 和 Pull Request。如有重大变更，请先开 Issue 讨论。

## 许可证

本项目基于 [Apache License 2.0](https://github.com/chenmortal/grafana-skywalking-datasource/blob/main/LICENSE) 许可。

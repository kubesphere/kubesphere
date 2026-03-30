# KubeSphere Skill 上线，为 AI 助手注入云原生能力


AI 助手越来越能干，但它真的了解你的平台吗？面对 KubeSphere 的扩展机制、API 规范和多集群架构，通用模型往往给不出准确答案。KubeSphere Skill 正式上线，把平台能力结构化地"教给"AI，让它真正用起来。


## 问题从哪里来

用过 AI 助手处理 Kubernetes 运维的人，大概都有过类似的体验：

问一个关于 KubeSphere 多集群管理的问题，AI 给了你一段通用的 kubectl 命令；问 DevOps 流水线怎么配置，它给你描述的是原生 Jenkins 的界面；问扩展组件怎么安装，它把 Helm Chart 的流程给你走了一遍——而 KubeSphere 的扩展管理根本不是这样运作的。

这不是模型"不够聪明"，而是模型**根本没有可靠的 KubeSphere 知识**。

通用大模型的训练数据里，KubeSphere 的比重非常有限，且可能已经滞后于当前版本。更关键的是，KubeSphere 有自己的架构设计、扩展机制和 API 规范——这些并不能从通用的 Kubernetes 知识里推断出来。

于是每次使用 AI 辅助 KubeSphere 相关工作，你都要花大量时间向它解释背景：这是什么平台、这个操作的正确姿势是什么、这个 API 的路径长什么样……效率远不如直接查文档。

**AI 成了"聪明但不熟悉业务的新人"，而你成了给它做培训的那个人。**


## Skill 要解决的，正是这个问题

近期，Anthropic 推动的 Agent Skills 规范正在成为 AI 助手扩展的事实标准。其核心思路是：与其每次对话都向 AI 重新描述一遍背景，不如把这些背景结构化地封装成可复用的"技能包"——让 AI 在需要时按需加载，而不是靠模型凭空猜测。

> Skill 将执行方法、工具调用方式以及相关知识材料，封装为一个完整的"能力扩展包"，使 AI 助手具备稳定、可复用的专业能力。

KubeSphere 在这个方向上给出了自己的答案：**kubesphere-skills**，一套专门为 AI 助手提供的 KubeSphere 技能库。

它的目标很直接：**让 AI 助手真正了解 KubeSphere 的架构、API 和操作方式**，从根本上消除 AI 在 KubeSphere 场景下的"知识盲区"。


## KubeSphere-skills 是什么

简单来说，kubesphere-skills 是一组结构化的知识包，覆盖 KubeSphere 平台的核心能力和扩展组件。每个 Skill 对应一个具体的功能域，告诉 AI 助手：这个模块是干什么的、架构是怎样的、正确的 API 用法是什么、常见操作怎么处理、典型问题怎么排查。

这些 Skill 的内容直接来源于 KubeSphere 官方代码库和文档，**保证了知识的准确性和及时性**，而不是依赖模型对公开资料的模糊记忆。

目前已包含的 Skill 涵盖两大方向：

### 核心平台层

| Skill | 描述 | 适用场景 |
|-------|------|---------|
| `kubesphere-core` | KubeSphere 核心平台架构 | 使用 KubeSphere 核心功能、扩展机制 |
| `kubesphere-extension-management` | 扩展组件生命周期管理 | 安装、配置、升级或排查扩展组件故障 |
| `kubesphere-cluster-management` | 集群查询（只读） | 查询集群列表、状态、详情、版本信息 |
| `kubesphere-multi-tenant-management` | 多租户管理 | 工作空间、命名空间、角色与访问控制 |

### DevOps 扩展层

| Skill | 描述 | 适用场景 |
|-------|------|---------|
| `kubesphere-devops-overview` | DevOps 扩展整体架构 | DevOps 搭建、架构理解、通用 CI/CD 操作 |
| `kubesphere-devops-pipeline` | 流水线管理 | 通过 API 创建、运行和监控 CI/CD 流水线 |
| `kubesphere-devops-credentials` | 凭据管理 | 管理仓库和部署凭据 |
| `kubesphere-devops-jenkins` | Jenkins 配置 | Jenkins 配置、代理管理、认证与故障排查 |
| `kubesphere-devops-tenant` | 租户操作规范 | 以命名空间范围租户身份运行 DevOps |
| `kubesphere-devops-argocd` | ArgoCD 集成 | GitOps 应用部署与管理 |



## 装上之后，体验有什么不同

同样的问题，加载 Skill 前后，AI 给出的答案截然不同：

| 场景 | ❌ 加载 Skill 之前 | ✅ 加载 Skill 之后 |
|------|------------------|------------------|
| **扩展组件安装** | 给出 `helm install` 命令和通用参数 | 正确使用 `InstallPlan` CRD，说明 name 必须匹配 extension.name，version 必须精确版本 |
| **集群信息查询** | 只会用 `kubectl get nodes` | 正确使用 `kubectl get cluster` 查询 KubeSphere 集群资源，知道如何判断 Host/Member 集群 |
| **创建企业空间及授权** | 只会用 `kubectl create` 手动创建资源 | 正确调用 KubeSphere API 创建 Workspace、Project、User，并按最小权限原则分配角色 |
| **DevOps 流水线触发** | 建议调用 Jenkins API `job/xxx/build` | 正确使用 `PipelineRun` CR 触发流水线，提供完整的 YAML 结构和 API 路径 |
| **GitOps 应用部署** | 建议直接创建 ArgoCD Application | 知道 KubeSphere 提供租户友好的 GitOps API，无需访问 argocd 命名空间即可创建应用 |


**AI 给出的，是基于 KubeSphere 真实规范的操作建议，而不是从通用知识推断出的近似答案。**


## 这件事为什么值得认真做

从技术角度看，Skill 的价值在于它把知识从"一次性对话输入"变成了"可持久复用的结构化资产"。对于 KubeSphere 这样一个有复杂扩展生态的平台，这一点尤为重要。

- **对个人用户：** 不用再反复给 AI"补课"，加载一次 Skill，AI 就拥有了平台知识，可以直接进入问题本身。

- **对团队：** 团队所有人共享同一套 Skill，AI 助手给出的建议基于统一的知识基础，而不是每个人自己调教出来的各种版本。

- **对平台本身：** Skill 直接来源于官方文档和代码库，随平台版本迭代更新。知识是有来源、可追溯、可维护的，不是模型凭空生成的。



## 现在就能用

kubesphere-skills 已开源发布，安装方式非常简单：

```bash
# 克隆仓库
git clone https://github.com/kubesphere/kubesphere-skills.git

# 将 Skill 复制到 AI 助手的 skills 目录，以 opencode 为例
cp -r kubesphere-skills/core/* ~/.config/opencode/skills/
cp -r kubesphere-skills/extension-devops/* ~/.config/opencode/skills/
```

完成后，你的 AI 助手就加载了 KubeSphere 的平台知识，可以直接投入使用。

社区贡献同样欢迎。如果你在使用中发现某个操作场景缺少对应的 Skill，或者现有 Skill 描述有改进空间，可以按照仓库的贡献指南提交——包括压力场景、API 示例和实际排查步骤。

📦 **项目地址：** [github.com/kubesphere/kubesphere-skills](https://github.com/kubesphere/kubesphere-skills)



## 写在最后

让 AI 助手真正在云原生场景里帮上忙，前提不是把提示词写得更精巧，而是把平台的知识结构化地告诉它。

kubesphere-skills 做的正是这件事：把 KubeSphere 的架构理解、API 规范、操作最佳实践，沉淀为 AI 可以加载、复用、可持续维护的技能包。

**AI 终于不用再猜了，KubeSphere 终于被真正理解了。**


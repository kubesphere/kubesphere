## Network API 参考

### IPPool

Kubesphere 3.5 之前使用 ippools.network.kubesphere.io 进行管理 ippool，再由 ks-controller-manager 通过 ippools.network.kubesphere.io 来间接管理 calico 的 ippools.crd.projectcalico.org。但是客户有可能会使用其他的运维平台来直接管理 calico 的 ippools.crd.projectcalico.org，由此不同的管理方式导致网络表现不符合预期，造成冲突。因此我们舍弃了 kubesphere 的 ippools.network.kubesphere.io，回退回直接管理 calico ippools.crd.projectcalico.org，这样可以避免不同管理方式造成的冲突

这样的改动也导致 api 发生以下的变化：

**创建、修改、删除 ippool**
- apis/crd.projectcalico.org/v1/ippools

**ippool ip 列表、占用详情**
- /kapis/network.kubesphere.io/v1alpha2/ippools
- /kapis/network.kubesphere.io/v1alpha2/ippools/{name}

**ippool 被绑定的项目列表**
- /kapis/resources.kubesphere.io/v1alpha3/namespaces?sortBy=createTime&limit=10&labelSelector=ippool.network.kubesphere.io%2Fippool-2

**ippool 容器组占用详情**
- /kapis/resources.kubesphere.io/v1alpha3/pods?limit=6&labelSelector=ippool.network.kubesphere.io%2Fname%3Ddefault-ipv4-ippool&sortBy=startTime

**namespace 绑定/取消绑定 ippool**
- 使用 patch or put 均可 /api/v1/namespaces/project

**ippool 取消全部 namespace 绑定信息**
- 获取 ippool 被绑定的 ns
- 逐个遍历 ns，使用 patch or put 取消绑定 /api/v1/namespaces/project

**迁移 ippool**
- /kapis/network.kubesphere.io/v1alpha2/ippoolmigrations
- body：oldippool=xxx newippool=xxx

**获取可以迁移的 ippool 列表**
- /kapis/network.kubesphere.io/v1alpha2/ippools/{name}/migrate

**获取 namespace 可用的 ippool 列表**
- /kapis/network.kubesphere.io/v1alpha2/namespaces/default/ippools


### NetworkPolicy

**集群视角下的网络策略管理**
- 查询：
	- /kapis/networking.k8s.io/v1/networkpolicies
    - /kapis/networking.k8s.io/v1/namespaces/{namespace}/networkpolicies?page=1&sortBy=createTime&limit=10
- 创建post /kapis/networking.k8s.io/v1/namespaces/project/networkpolicies
- 删除del /kapis/networking.k8s.io/v1/namespaces/project/networkpolicies/{name}

**企业空间视角下的网络策略管理**
- 判断是否启用：
	根据 workspace annotation 中 kubesphere.io/workspace-isolate 的值判断；enabled 为启用；不存在或其他值为未启用
- 启用/禁用 patch or uprate /apis/tenant.kubesphere.io/v1beta1/workspaces/{name}
```yaml
{
  "metadata": {
    "annotations": {
      "kubesphere.io/network-isolate": "enabled"
    }
  }
}
```

**项目视角下的网络策略管理**
- 判断是否启用
	根据 namespace annotation 中 kubesphere.io/workspace-isolate 的值判断；enabled 为启用；不存在或其他值为未启用
- 启用/禁用 patch namespace： /api/v1/namespaces/{namespace}
```yaml
{
  "metadata": {
    "annotations": {
      "kubesphere.io/network-isolate": "enabled"
    }
  }
}
```
- 查询：Get /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies?sortBy=createTime&limit=10
- 创建： Post  /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies
- 删除： Delete /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies/{name}
- 编辑：Put /kapis/network.kubesphere.io/v1alpha1/namespaces/{namespace}/namespacenetworkpolicies/{name}

创建时注意：
- 创建内部白名单时指定 labels：kubesphere.io/policy-traffic=inside
- 创建外部白名单时指定 labels：kubesphere.io/policy-traffic=outside
- 创建出站白名单时指定 labels：kubesphere.io/policy-type=egress
- 创建入站白名单时指定 labels：kubesphere.io/policy-type=ingress

创建指定外部白名单的出站流量：
```yaml
    "labels": {
      "kubesphere.io/policy-type": "egress",
      "kubesphere.io/policy-traffic": "outside"
    },
```
查询指定外部白名单的出站流量：
/kapis/network.kubesphere.io/v1alpha1/namespaces/project-2/namespacenetworkpolicies?page=1&sortBy=createTime&limit=10&labelSelector=kubesphere.io%2Fpolicy-type%3Degress%2Ckubesphere.io%2Fpolicy-traffic%3Doutside


## 项目说明

用 kubebuilder 生成的框架代码

实现了对自定义资源类型 CustomPod 跟 Deployment 一样的副本控制

CustomPod实际副本数比期望(yaml)的少，就创建，比期望的多，就删除多余的副本

修改了 CRD 定义：custompod_types.go 和 控制器 custompod_controller.go 的代码

### 启动配置

config/default 在标准配置中包含 Kustomize base ，它用于启动控制器。

CRD yaml：config/crd/bases/zhhnzw.mock.com_custompods.yaml

CRD 实例 CR 的 yaml：config/samples/zhhnzw_v1_custompod.yaml

config/manager: 在集群中以 pod 的形式启动控制器（部署在集群中时）

config/rbac: 在自己的账户下运行控制器所需的权限

### 这个项目是如何初始化的？

```bash
$ kubebuilder init --domain mock.com --repo github.com/zhhnzw/k8s-demo/kubebuilder-demo/v1
$ kubebuilder create api --group zhhnzw --version v1 --kind=CustomPod --resource=true --controller=true
```

### 启动

```bash
# 更新 crd/bases yaml
$ make manifests

# 更新 zz_generated.deepcopy.go
$ make generate

# Install the CRDs into the cluster:
$ make install

# Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):
$ make run
```

### 功能验证

```bash
# 发布一个crd实例
$ kubectl apply -f config/samples/zhhnzw_v1_custompod.yaml

# 查看pod，看是否如期望创建了1个pod
$ kubectl get pods

# 修改 sample yaml 文件的 replicas 为 3
$ kubectl apply -f config/samples/zhhnzw_v1_custompod.yaml

# 查看pod，看是否如期望的把pod副本数量增加到了3个
$ kubectl get pods

# 修改 sample yaml 文件的 replicas 为 1
$ kubectl apply -f config/samples/zhhnzw_v1_custompod.yaml

# 查看pod，看是否如期望的把pod副本数量减少到了1个
$ kubectl get pods
```
### 准备工作
挂载 RAM Disk（内存盘）来模拟本地磁盘，在Linux下的命令示例：
```bash
# 在 node-1 上执行
$ mkdir /mnt/disks
$ for vol in vol1 vol2 vol3; do
    mkdir /mnt/disks/$vol
    mount -t tmpfs $vol /mnt/disks/$vol
done
```

mac 不支持 tmpfs，可用 diskutil，

RAM disk 把内存当硬盘使，推出该 Ramdisk 即可释放内存，然后绑定在本地磁盘地址的数据也会全部丢失。

格式如下：

```bash
diskutil erasevolume HFS+ "<名称>" `hdiutil attach -nomount ram://$((<容量（GB）>*2097152))`
```

例如：
```bash
$ diskutil erasevolume HFS+ "k8s-disk" `hdiutil attach -nomount ram://$((1*2097152))`
Started erase on disk2
Unmounting disk
Erasing
Initialized /dev/rdisk2 as a 1024 MB case-insensitive HFS Plus volume
Mounting disk
Finished erase on disk2 (k8s-disk)
```

操作成功后的挂载路径：/Volumes/k8s-disk

创建 k8s node 标签，指令格式：
```bash
$ kubectl label nodes <node-name> <label-key>=<label-value> 
```
例如：
```bash
$ kubectl get nodes
NAME             STATUS   ROLES    AGE    VERSION
docker-desktop   Ready    master   153d   v1.19.7
# 可以看到有多组标签，可以找到kubernetes.io/hostname=docker-desktop
# 可以找到InternalIP:192.168.65.4
$ kubectl describe node docker-desktop 
$ kubectl label node 192.168.65.4 custom_label=test # 给指定node添加自定义标签
```

### 创建pv、storageClass、pvc
```bash
$ kubectl create -f local_pv.yaml 
persistentvolume/example-pv created
 
$ kubectl get pv
NAME         CAPACITY   ACCESS MODES   RECLAIM POLICY  STATUS      CLAIM             STORAGECLASS    REASON    AGE
example-pv   5Gi        RWO            Delete           Available  

$ kubectl create -f local_sc.yaml 
storageclass.storage.k8s.io/local-storage created

$ kubectl create -f local_pvc.yaml 
persistentvolumeclaim/example-local-claim created
 
$ kubectl get pvc
NAME                  STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS    AGE
example-local-claim   Pending 
  
```

### 创建Pod
```bash
$ kubectl create -f local_pod.yaml 
pod/example-pv-pod created
 
$ kubectl get pvc # 创建pod后STATUS就变成了Bound
NAME                  STATUS    VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS    AGE
example-local-claim   Bound     example-pv   5Gi        RWO            local-storage   6h
```

### 测试localPV的功能
进入容器创建文件
```bash
$ kubectl exec -it example-pv-pod -- /bin/sh
# cd /usr/share/nginx/html  # Pod 把volume挂载到了 /usr/share/nginx/html 
# touch test.txt
```
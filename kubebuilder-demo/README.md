```bash
# 更新 crd yaml
$ make manifests

# 更新 zz_generated.deepcopy.go
$ make generate

# Install the CRDs into the cluster:
$ make install

# Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):
$ make run

# Install Instances of Custom Resources
$ kubectl apply -f config/samples/

# 查看创建的Custom Resources
$ kubectl get pods
```


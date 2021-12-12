## 代码添加依赖 
github.com/prometheus/client_golang 
prometheus go客户端：支持prometheus通过/metrics接口采集数据

## prometheus安装

利用prometheus-operator安装：
地址：https://github.com/prometheus-operator/kube-prometheus

```shell
#克隆代码
git clone https://github.com/prometheus-operator/kube-prometheus.git
cd manifests
#查看安装资源清单
ls
kubectl apply -f .
kubectl get pods -n monitoring
#查看创建的service
kubectl get svc -n monitoring

#为了方便k8s集群外访问，编辑 grafana 和 prometheus-k8s 这两个 Service，将服务类型更改为 NodePort
kubectl edit svc grafana -n monitoring
kubectl edit svc prometheus-k8s -n monitoring
```

## prometheus配置 自动发现配置
配置自动发现，可以再pod上添加注解，prometheus可自动发现并进行注册
annotations:
prometheus.io/scrape: "true"
prometheus.io/port: "8080"

当然也可以添加其他注解，比如一些业务注解。

## 配置grafana

查找相关dashboards，获取编号，进行导入，有问题再进行修改。
https://grafana.com/grafana/dashboards/
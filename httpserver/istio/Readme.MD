把httpserver 服务以 Istio Ingress Gateway 的形式发布出来:
如何实现安全保证；
七层路由规则；
考虑 open tracing 的接入。

### 创建tls证书
```shell
# 创建证书信息
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=lendea.com/O=lendea"

openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=lendea Inc./CN=*.lendea.com' -keyout lendea.com.key -out lendea.com.crt
# 创建k8s secret 保存证书信息
kubectl create -n istio-system secret tls lendea-credential --key=lendea.com.key --cert=lendea.com.crt
```


### 安装链路追踪服务
``` shell
# 安装jaeger
kubectl apply -f jaeger.yaml
# 调整istio链路追踪上报数据的采样率
kubectl edit configmap istio -n istio-system
```

```yaml
  tracing:
    sampling: 100
```


### 部署httpserver
```shell
# 创建namespace
kubectl create ns tracing
# 开启istio注入
kubectl label ns tracing istio-injection=enabled
# 部署httpserver
kubectl -n tracing apply -f deploy.yaml
# 部署istio配置
kubectl apply -f istio.yaml -n tracing
```
通过istio-gateway使用https的方式统一暴露httpserver，保障入口安全
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: httpsserver
spec:
  gateways:
    - httpsserver
  hosts:
    - httpsserver.lendea.com
  http:
    # 转发规则，访问443端口时，转发到httpserver-svc的80
    - match:
        - port: 443
      route:
        - destination:
            host: httpserver-svc.tracing.svc.cluster.local
            port:
              number: 80
---
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: httpsserver
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - httpsserver.lendea.com
      port:
        name: https-default
        number: 443
        protocol: HTTPS
      # 引用前面创建的tls证书  
      tls:
        mode: SIMPLE
        credentialName: lendea-credential
```



### 检查istio的ingressgateway服务的IP信息
```shell
kubectl get svc -nistio-system
```
返回的ingressgateway信息中包含IP信息
```text
istio-ingressgateway   LoadBalancer  10.101.155.80
```


### 访问入口网关
```shell
 curl --resolve httpsserver.lendea.com:443:10.101.155.80 https://httpsserver.lendea.com/teset  -v -k
```

### 开启jaeger链路追踪服务的dashboard
```shell
istioctl dashboard jaeger
```
访问http://localhost:16686/jaeger/search，查看服务访问的链路信息



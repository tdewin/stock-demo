# Stock app demo
Buy some fake produce. Mainly to manage some stateful data in postgres DB

First create a namespace
```
kubectl create ns stock-demo
```

Then install postgres via helm for example
```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install -n stock-demo --set global.postgresql.auth.username=root --set global.postgresql.auth.password=notsecure --set global.postgresql.auth.database=stock stockdb bitnami/postgresql
```

Then deploy the app. Service deploys as LoadBalancer
```
kubectl -n stock-demo apply -f https://raw.githubusercontent.com/tdewin/stock-demo/main/kubernetes/deployment.yaml
kubectl -n stock-demo apply -f https://raw.githubusercontent.com/tdewin/stock-demo/main/kubernetes/svc.yaml
```
After deployment, use /init to create the basic table and insert some fake produce
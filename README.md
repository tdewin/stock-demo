# Stock app demo
Buy some fake produce. Mainly to manage some stateful data in postgres DB

First install postgres via helm for example
```
kubectl create ns stock-demo
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install -n stock-demo --set global.postgresql.auth.username=root --set global.postgresql.auth.password=notsecure --set global.postgresql.auth.database=stock stockdb bitnami/postgresql
```

After deployment, use /init to create the basic table and insert some fake produce
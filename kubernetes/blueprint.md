Matching annotation for blueprint
```
kubectl annotate statefulset stockdb-postgresql kanister.kasten.io/blueprint='postgresql-application-consistent' --namespace=stock-demo
```

Extracted with
```
kubectl get blueprint -n kasten-io  postgresql-application-consistent -o yaml
```

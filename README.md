# README


```m̀ermaids
graph LR
    A[apply] --> B[kubectl apply -f <file>]
    A --> C[kubectl create -f <file>]
    A --> D[kubectl edit -f <file>]
    B --o[options] Y{
    Y1[--dry-run=client/server]
    Y2[--namespace=<namespace>]
    }
    C --o[options] Z{
    Z1[--dry-run=client/server]
    Z2[--force]
    Z3[--namespace=<namespace>]
    }
    D --o[options] AA{
    AA1[--dry-run=client/server]
    AA2[--editor=<editor>]
    }
    E[delete] --> F[kubectl delete -f <file>]
    F --o[options] BB{
    BB1[--grace-period=<duration>]
    BB2[--force]
    BB3[--ignore-not-found]
    BB4[--namespace=<namespace>]
    }
    G[get] --> H[kubectl get <resource> [<name>|-o wide]]
    H --o[options] CC{
    CC1[--dry-run=client/server]
    CC2[--output=json/yaml/template/wide/custom-columns/go-template/name/raw]
    CC3[--watch]
    }
    I[describe] --> J[kubectl describe <resource> <name>]
    K[logs] --> L[kubectl logs <pod-name> [-f]]

```
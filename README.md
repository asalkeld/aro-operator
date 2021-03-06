# Azure Red Hat OpenShift Operator

## MVP

* monitor and repair pull secret (acr part)
* periodically check for internet connectivity and mark as supported/unsupported
* monitor and repair mdsd as needed

## Future responsibilities

### Decentralizing service monitoring

* periodically check for internet connectivity and mark as supported/unsupported

### Automatic service remediation

* monitor and repair pull secret (acr part)
* monitor and repair mdsd as needed

### End user warnings

### Decentralizing ARO customization management

* take over install customizations

## dev help

```sh
export OPERATOR_NAME=aro-operator
operator-sdk run --local --kubeconfig $KUBECONFIG --watch-namespace=""
```

see: https://sdk.operatorframework.io/docs/golang/quickstart/

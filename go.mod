module github.com/asalkeld/aro-operator

go 1.13

require (
	github.com/Azure/ARO-RP v0.0.0-20200507154943-98a37a303b96
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/openshift/api v0.0.0-20200429152225-b98a784d8e6d
	github.com/operator-framework/operator-sdk v0.17.1-0.20200508235800-4e2c562a3d29
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)

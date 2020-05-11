package deploy

import (
	"errors"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
)

func OperatorNamespace() (string, error) {
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		if errors.Is(err, k8sutil.ErrRunLocal) {
			return "default", nil
		}
		return "", err
	}
	return operatorNs, nil
}

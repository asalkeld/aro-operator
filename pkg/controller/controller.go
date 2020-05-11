package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, <-chan struct{}) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, stopCh <-chan struct{}) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, stopCh); err != nil {
			return err
		}
	}
	return nil
}

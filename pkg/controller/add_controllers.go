package controller

import (
	"github.com/asalkeld/aro-operator/pkg/controller/cluster"
	"github.com/asalkeld/aro-operator/pkg/controller/internetchecker"
	"github.com/asalkeld/aro-operator/pkg/controller/pullsecret"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, pullsecret.Add)
	AddToManagerFuncs = append(AddToManagerFuncs, internetchecker.Add)
	AddToManagerFuncs = append(AddToManagerFuncs, cluster.Add)
}

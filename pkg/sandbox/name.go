package sandbox

import (
	"fmt"
	"strings"

	devopsv1 "github.com/stackvista/sandbox-operator/apis/devops/v1"
)

// Construct the name of the actual Sandbox that is created based on the CRD
func SandboxName(sandbox *devopsv1.Sandbox) string {
	name := "sandbox"
	if !strings.HasPrefix(sandbox.Name, sandbox.Spec.User) {
		name = fmt.Sprintf("%s-%s", name, sandbox.Spec.User)
	}

	name = fmt.Sprintf("%s-%s", name, sandbox.Name)

	return name

}

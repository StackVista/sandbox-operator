/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"

	"gitlab.com/stackvista/devops/devopserator/cmd"
	logr "gitlab.com/stackvista/devops/devopserator/internal/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

func main() {
	logger := zap.New(zap.UseDevMode(true))
	ctx := context.Background()

	ctx = logr.WithContext(ctx, logger)

	cmd.Execute(ctx)
}

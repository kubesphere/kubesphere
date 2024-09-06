/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package main

import (
	"os"

	"k8s.io/component-base/cli"

	"kubesphere.io/kubesphere/cmd/controller-manager/app"
)

func main() {
	cmd := app.NewControllerManagerCommand()
	code := cli.Run(cmd)
	os.Exit(code)
}

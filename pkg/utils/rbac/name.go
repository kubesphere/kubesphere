/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package rbac

import "fmt"

const iamPrefix = "kubesphere:iam"

func RelatedK8sResourceName(name string) string {
	return fmt.Sprintf("%s:%s", iamPrefix, name)
}

/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package net

// 0 is considered as a non valid port
func IsValidPort(port int) bool {
	return port > 0 && port < 65535
}

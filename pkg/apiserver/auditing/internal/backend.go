/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package internal

type Backend interface {
	ProcessEvents(events ...[]byte)
}

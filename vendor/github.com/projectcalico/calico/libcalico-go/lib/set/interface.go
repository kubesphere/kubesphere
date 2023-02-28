// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package set

import (
	"errors"
	"fmt"
)

type Set[T any] interface {
	Len() int
	Add(T)
	AddAll(itemArray []T)
	AddSet(other Set[T])
	Discard(T)
	Clear()
	Contains(T) bool
	Iter(func(item T) error)
	Copy() Set[T]
	Equals(Set[T]) bool
	ContainsAll(Set[T]) bool
	Slice() []T
	fmt.Stringer
}

var (
	StopIteration = errors.New("stop iteration")
	RemoveItem    = errors.New("remove item")
)

type v struct{}

var emptyValue = v{}

/*
Open Source Initiative OSI - The MIT License (MIT):Licensing

The MIT License (MIT)
Copyright (c) 2013 Ralph Caraveo (deckarep@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mapset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type ThreadUnsafeSet map[interface{}]struct{}

// An OrderedPair represents a 2-tuple of values.
type OrderedPair struct {
	First  interface{}
	Second interface{}
}

func newThreadUnsafeSet() ThreadUnsafeSet {
	return make(ThreadUnsafeSet)
}

// Equal says whether two 2-tuples contain the same values in the same order.
func (pair *OrderedPair) Equal(other OrderedPair) bool {
	if pair.First == other.First &&
		pair.Second == other.Second {
		return true
	}

	return false
}

func (set *ThreadUnsafeSet) Add(i interface{}) bool {
	_, found := (*set)[i]
	if found {
		return false //False if it existed already
	}

	(*set)[i] = struct{}{}
	return true
}

func (set *ThreadUnsafeSet) Contains(i ...interface{}) bool {
	for _, val := range i {
		if _, ok := (*set)[val]; !ok {
			return false
		}
	}
	return true
}

func (set *ThreadUnsafeSet) IsSubset(other Set) bool {
	_ = other.(*ThreadUnsafeSet)
	if set.Cardinality() > other.Cardinality() {
		return false
	}
	for elem := range *set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set *ThreadUnsafeSet) IsProperSubset(other Set) bool {
	return set.IsSubset(other) && !set.Equal(other)
}

func (set *ThreadUnsafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set *ThreadUnsafeSet) IsProperSuperset(other Set) bool {
	return set.IsSuperset(other) && !set.Equal(other)
}

func (set *ThreadUnsafeSet) Union(other Set) Set {
	o := other.(*ThreadUnsafeSet)

	unionedSet := newThreadUnsafeSet()

	for elem := range *set {
		unionedSet.Add(elem)
	}
	for elem := range *o {
		unionedSet.Add(elem)
	}
	return &unionedSet
}

func (set *ThreadUnsafeSet) Intersect(other Set) Set {
	o := other.(*ThreadUnsafeSet)

	intersection := newThreadUnsafeSet()
	// loop over smaller set
	if set.Cardinality() < other.Cardinality() {
		for elem := range *set {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range *o {
			if set.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return &intersection
}

func (set *ThreadUnsafeSet) Difference(other Set) Set {
	_ = other.(*ThreadUnsafeSet)

	difference := newThreadUnsafeSet()
	for elem := range *set {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return &difference
}

func (set *ThreadUnsafeSet) SymmetricDifference(other Set) Set {
	_ = other.(*ThreadUnsafeSet)

	aDiff := set.Difference(other)
	bDiff := other.Difference(set)
	return aDiff.Union(bDiff)
}

func (set *ThreadUnsafeSet) Clear() {
	*set = newThreadUnsafeSet()
}

func (set *ThreadUnsafeSet) Remove(i interface{}) {
	delete(*set, i)
}

func (set *ThreadUnsafeSet) Cardinality() int {
	return len(*set)
}

func (set *ThreadUnsafeSet) Each(cb func(interface{}) bool) {
	for elem := range *set {
		if cb(elem) {
			break
		}
	}
}

func (set *ThreadUnsafeSet) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		for elem := range *set {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

func (set *ThreadUnsafeSet) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
	L:
		for elem := range *set {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
	}()

	return iterator
}

func (set *ThreadUnsafeSet) Equal(other Set) bool {
	_ = other.(*ThreadUnsafeSet)

	if set.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range *set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set *ThreadUnsafeSet) Clone() Set {
	clonedSet := newThreadUnsafeSet()
	for elem := range *set {
		clonedSet.Add(elem)
	}
	return &clonedSet
}

func (set *ThreadUnsafeSet) String() string {
	items := make([]string, 0, len(*set))

	for elem := range *set {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(items, ", "))
}

// String outputs a 2-tuple in the form "(A, B)".
func (pair OrderedPair) String() string {
	return fmt.Sprintf("(%v, %v)", pair.First, pair.Second)
}

func (set *ThreadUnsafeSet) Pop() interface{} {
	for item := range *set {
		delete(*set, item)
		return item
	}
	return nil
}

func (set *ThreadUnsafeSet) PowerSet() Set {
	powSet := NewThreadUnsafeSet()
	nullset := newThreadUnsafeSet()
	powSet.Add(&nullset)

	for es := range *set {
		u := newThreadUnsafeSet()
		j := powSet.Iter()
		for er := range j {
			p := newThreadUnsafeSet()
			if reflect.TypeOf(er).Name() == "" {
				k := er.(*ThreadUnsafeSet)
				for ek := range *(k) {
					p.Add(ek)
				}
			} else {
				p.Add(er)
			}
			p.Add(es)
			u.Add(&p)
		}

		powSet = powSet.Union(&u)
	}

	return powSet
}

func (set *ThreadUnsafeSet) CartesianProduct(other Set) Set {
	o := other.(*ThreadUnsafeSet)
	cartProduct := NewThreadUnsafeSet()

	for i := range *set {
		for j := range *o {
			elem := OrderedPair{First: i, Second: j}
			cartProduct.Add(elem)
		}
	}

	return cartProduct
}

func (set *ThreadUnsafeSet) ToSlice() []interface{} {
	keys := make([]interface{}, 0, set.Cardinality())
	for elem := range *set {
		keys = append(keys, elem)
	}

	return keys
}

// MarshalJSON creates a JSON array from the set, it marshals all elements
func (set *ThreadUnsafeSet) MarshalJSON() ([]byte, error) {
	items := make([]string, 0, set.Cardinality())

	for elem := range *set {
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}

		items = append(items, string(b))
	}

	return []byte(fmt.Sprintf("[%s]", strings.Join(items, ","))), nil
}

// UnmarshalJSON recreates a set from a JSON array, it only decodes
// primitive types. Numbers are decoded as json.Number.
func (set *ThreadUnsafeSet) UnmarshalJSON(b []byte) error {
	var i []interface{}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	for _, v := range i {
		switch t := v.(type) {
		case []interface{}, map[string]interface{}:
			continue
		default:
			set.Add(t)
		}
	}

	return nil
}

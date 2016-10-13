// autogenerated: do not edit!
// generated from gentemplate [gentemplate -d Package=mctree -id shared_pair_offsets_pool -d PoolType=shared_pair_offsets_pool -d Type=shared_pair_offsets -d Data=elts github.com/platinasystems/go/elib/pool.tmpl]

package mctree

import (
	"github.com/platinasystems/go/elib"
)

type shared_pair_offsets_pool struct {
	elib.Pool
	elts []shared_pair_offsets
}

func (p *shared_pair_offsets_pool) GetIndex() (i uint) {
	l := uint(len(p.elts))
	i = p.Pool.GetIndex(l)
	if i >= l {
		p.Validate(i)
	}
	return i
}

func (p *shared_pair_offsets_pool) PutIndex(i uint) (ok bool) {
	return p.Pool.PutIndex(i)
}

func (p *shared_pair_offsets_pool) IsFree(i uint) (v bool) {
	v = i >= uint(len(p.elts))
	if !v {
		v = p.Pool.IsFree(i)
	}
	return
}

func (p *shared_pair_offsets_pool) Resize(n uint) {
	c := elib.Index(cap(p.elts))
	l := elib.Index(len(p.elts) + int(n))
	if l > c {
		c = elib.NextResizeCap(l)
		q := make([]shared_pair_offsets, l, c)
		copy(q, p.elts)
		p.elts = q
	}
	p.elts = p.elts[:l]
}

func (p *shared_pair_offsets_pool) Validate(i uint) {
	c := elib.Index(cap(p.elts))
	l := elib.Index(i) + 1
	if l > c {
		c = elib.NextResizeCap(l)
		q := make([]shared_pair_offsets, l, c)
		copy(q, p.elts)
		p.elts = q
	}
	if l > elib.Index(len(p.elts)) {
		p.elts = p.elts[:l]
	}
}

func (p *shared_pair_offsets_pool) Elts() uint {
	return uint(len(p.elts)) - p.FreeLen()
}

func (p *shared_pair_offsets_pool) Len() uint {
	return uint(len(p.elts))
}

func (p *shared_pair_offsets_pool) Foreach(f func(x shared_pair_offsets)) {
	for i := range p.elts {
		if !p.Pool.IsFree(uint(i)) {
			f(p.elts[i])
		}
	}
}
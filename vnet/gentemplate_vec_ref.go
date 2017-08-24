// autogenerated: do not edit!
// generated from gentemplate [gentemplate -d Package=vnet -id Ref -d VecType=RefVec -d Type=Ref github.com/platinasystems/go/elib/vec.tmpl]

// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
)

type RefVec []Ref

func (p *RefVec) Resize(n uint) {
	c := uint(cap(*p))
	l := uint(len(*p)) + n
	if l > c {
		c = elib.NextResizeCap(l)
		q := make([]Ref, l, c)
		copy(q, *p)
		*p = q
	}
	*p = (*p)[:l]
}

func (p *RefVec) validate(new_len uint, zero Ref) *Ref {
	c := uint(cap(*p))
	lʹ := uint(len(*p))
	l := new_len
	if l <= c {
		// Need to reslice to larger length?
		if l > lʹ {
			*p = (*p)[:l]
			for i := lʹ; i < l; i++ {
				(*p)[i] = zero
			}
		}
		return &(*p)[l-1]
	}
	return p.validateSlowPath(zero, c, l, lʹ)
}

func (p *RefVec) validateSlowPath(zero Ref, c, l, lʹ uint) *Ref {
	if l > c {
		cNext := elib.NextResizeCap(l)
		q := make([]Ref, cNext, cNext)
		copy(q, *p)
		for i := c; i < cNext; i++ {
			q[i] = zero
		}
		*p = q[:l]
	}
	if l > lʹ {
		*p = (*p)[:l]
	}
	return &(*p)[l-1]
}

func (p *RefVec) Validate(i uint) *Ref {
	var zero Ref
	return p.validate(i+1, zero)
}

func (p *RefVec) ValidateInit(i uint, zero Ref) *Ref {
	return p.validate(i+1, zero)
}

func (p *RefVec) ValidateLen(l uint) (v *Ref) {
	if l > 0 {
		var zero Ref
		v = p.validate(l, zero)
	}
	return
}

func (p *RefVec) ValidateLenInit(l uint, zero Ref) (v *Ref) {
	if l > 0 {
		v = p.validate(l, zero)
	}
	return
}

func (p *RefVec) ResetLen() {
	if *p != nil {
		*p = (*p)[:0]
	}
}

func (p RefVec) Len() uint { return uint(len(p)) }

// autogenerated: do not edit!
// generated from gentemplate [gentemplate -id adjAddDelHook -d Package=ip -d DepsType=adjAddDelHookVec -d Type=adjAddDelHook -d Data=adjAddDelHooks github.com/platinasystems/go/elib/dep/dep.tmpl]

package ip

import (
	"github.com/platinasystems/go/elib/dep"
)

type adjAddDelHookVec struct {
	deps           dep.Deps
	adjAddDelHooks []adjAddDelHook
}

func (t *adjAddDelHookVec) Len() int {
	return t.deps.Len()
}

func (t *adjAddDelHookVec) Get(i int) adjAddDelHook {
	return t.adjAddDelHooks[t.deps.Index(i)]
}

func (t *adjAddDelHookVec) Add(x adjAddDelHook, ds ...*dep.Dep) {
	if len(ds) == 0 {
		t.deps.Add(&dep.Dep{})
	} else {
		t.deps.Add(ds[0])
	}
	t.adjAddDelHooks = append(t.adjAddDelHooks, x)
}
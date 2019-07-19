package rplx

import (
	"github.com/pkg/errors"
)

var (
	ErrVariableNotExists = errors.New("variable not exists")
)

// get returns variable v or error if variable not exists
func (rplx *Rplx) Get(name string) (int64, error) {
	rplx.variablesMx.RLock()
	defer rplx.variablesMx.RUnlock()

	v, ok := rplx.variables[name]
	if !ok {
		return 0, ErrVariableNotExists
	}

	return v.get()
}

// Upsert change variable on delta or create variable, if not exists
func (rplx *Rplx) Upsert(name string, delta int64) {
	var v *variable
	var ok bool

	rplx.variablesMx.RLock()
	v, ok = rplx.variables[name]
	rplx.variablesMx.RUnlock()
	if !ok {
		rplx.variablesMx.Lock()
		v, ok = rplx.variables[name]
		if !ok {
			v = newVariable(name)
			rplx.variables[name] = v
		}
		rplx.variablesMx.Unlock()
	}

	v.update(delta)

	go rplx.sendToReplication(v)
}

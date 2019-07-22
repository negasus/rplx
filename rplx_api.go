package rplx

import (
	"github.com/pkg/errors"
	"time"
)

var (
	ErrVariableNotExists = errors.New("variable not exists")
)

// get returns variable v or error if variable not exists or expired
// if variable expired, removes variable from rplx.variables map
func (rplx *Rplx) Get(name string) (int64, error) {
	rplx.variablesMx.RLock()
	v, ok := rplx.variables[name]
	rplx.variablesMx.RUnlock()

	if !ok {
		return 0, ErrVariableNotExists
	}

	ttl := v.getTTL()

	if ttl > 0 && ttl < time.Now().UTC().UnixNano() {
		rplx.variablesMx.Lock()
		delete(rplx.variables, name)
		rplx.variablesMx.Unlock()

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

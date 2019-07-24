package rplx

import (
	"github.com/pkg/errors"
	"time"
)

var (
	// ErrVariableNotExists describe error for non exists variable
	ErrVariableNotExists = errors.New("variable not exists")
	// ErrTTLLessThanNow describe error for func UpdateTTL
	ErrTTLLessThanNow = errors.New("TTL less than Now")
)

// Get returns variable v or error if variable not exists or expired
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

// Delete sets for variable ttl with -1 sec from Now,
// sends variable to replication (update TTL on clients) and remove variable from rplx.variables map
func (rplx *Rplx) Delete(name string) error {
	rplx.variablesMx.Lock()
	defer rplx.variablesMx.Unlock()

	v, ok := rplx.variables[name]
	if !ok {
		return ErrVariableNotExists
	}

	v.updateTTL(time.Now().UTC().Add(-time.Second))

	go rplx.sendToReplication(v)

	delete(rplx.variables, name)

	return nil
}

// UpdateTTL updates TTL for variable or return error if variable not exists
func (rplx *Rplx) UpdateTTL(name string, ttl time.Time) error {

	if ttl.Before(time.Now().UTC()) {
		return ErrTTLLessThanNow
	}

	rplx.variablesMx.RLock()
	defer rplx.variablesMx.RUnlock()

	v, ok := rplx.variables[name]
	if !ok {
		return ErrVariableNotExists
	}

	v.updateTTL(ttl)

	go rplx.sendToReplication(v)

	return nil
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

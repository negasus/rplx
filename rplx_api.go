package rplx

import (
	"github.com/pkg/errors"
	"time"
)

var (
	// ErrVariableNotExists returns if variable not exists
	ErrVariableNotExists = errors.New("variable not exists")
	ErrVariableExpired   = errors.New("variable expired")
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

	ttl := v.TTL()

	if ttl > 0 && ttl < time.Now().UTC().UnixNano() {
		rplx.variablesMx.Lock()
		delete(rplx.variables, name)
		rplx.variablesMx.Unlock()

		return 0, ErrVariableExpired
	}

	return v.get(), nil
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

	v.updateTTL(time.Now().UTC().Add(-time.Second).UnixNano())

	go rplx.sendToReplication(v)

	delete(rplx.variables, name)

	return nil
}

// UpdateTTL updates TTL for variable or return error if variable not exists
func (rplx *Rplx) UpdateTTL(name string, ttl time.Time) error {
	rplx.variablesMx.RLock()
	defer rplx.variablesMx.RUnlock()

	v, ok := rplx.variables[name]
	if !ok {
		return ErrVariableNotExists
	}

	v.updateTTL(ttl.UnixNano())

	go rplx.sendToReplication(v)

	return nil
}

// Upsert change variable on delta or create variable, if not exists
// returns new value
func (rplx *Rplx) Upsert(name string, delta int64) int64 {
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

	ttl := v.TTL()
	// if variable has TTL and TTL less than Now, variable was expired, but not garbage collected
	if ttl > 0 && ttl < time.Now().UTC().UnixNano() {
		v.updateTTL(0)
		delta = delta - v.get()
	}

	v.update(delta)

	go rplx.sendToReplication(v)

	return v.get()
}

// All returns all variables values
// first returns param - not expires variables
// second param - expires, but not garbage collected variables
func (rplx *Rplx) All() (notExpired map[string]int64, expired map[string]int64) {
	notExpired = make(map[string]int64)
	expired = make(map[string]int64)

	rplx.variablesMx.RLock()
	defer rplx.variablesMx.RUnlock()

	for name, v := range rplx.variables {
		ttl := v.TTL()

		if ttl > 0 && ttl < time.Now().UTC().UnixNano() {
			expired[name] = v.get()
			continue
		}

		notExpired[name] = v.get()
	}

	return
}

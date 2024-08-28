package gui

import (
	"fmt"
	"testing"
)

// Kill the killer with no victim attached.
func TestKillerKill(t *testing.T) {
	killer := newKiller()
	if !trySend(killer.Kill(), true, timeout) {
		t.Errorf("kill timed out after %v", timeout)
	}
	if _, ok := tryRecv(killer.Dead(), timeout); !ok {
		t.Errorf("no dead signal from killer after %v", timeout)
	}
}

// Kill the killer with a victim attached.
func TestKillerAttachKill(t *testing.T) {
	killer := newKiller()
	victim, err := newDummyVictim(killer)
	if err != nil {
		t.Error(err)
	}

	// Kill the killer.
	if !trySend(killer.Kill(), true, timeout) {
		t.Errorf("failed to kill killer after %v", timeout)
	}
	if _, ok := tryRecv(killer.Dead(), timeout); !ok {
		t.Errorf("killer not dead after %v", timeout)
	}

	// victim.Dead() should now be closed.
	if _, notClosed := <-victim.Dead(); notClosed {
		t.Errorf("victim not dead after killing killer")
	}
}

// Detach the victim and attach another in its place.
func TestKillerReattach(t *testing.T) {
	killer := newKiller()

	// Attach first victim.
	victim1, err := newDummyVictim(killer)
	if err != nil {
		t.Error(err)
	}

	// Try to attach second victim while first still attached—should fail.
	if _, err := newDummyVictim(killer); err == nil {
		t.Errorf("killer accepted another victim while the first was still attached.")
	}

	// Detach first victim.
	if !trySend(victim1.Kill(), true, timeout) {
		t.Errorf("failed to kill first victim after %v", timeout)
	}
	if _, ok := tryRecv(victim1.Dead(), timeout); !ok {
		t.Errorf("first victim failed to die after %v", timeout)
	}

	// Attach second victim.
	if _, err := newDummyVictim(killer); err != nil {
		t.Error(err)
	}

	killer.Kill() <- true
	<-killer.Dead()
}

type dummyVictim struct {
	kill       chan<- bool
	dead       <-chan bool
	detachChan <-chan bool
}

// newDummyVictim returns a victim that is attached to parent,
// or error if the parent does not accept the attach.
func newDummyVictim(parent killer) (victim, error) {
	kill := make(chan bool)
	dead := make(chan bool)
	detachChan := make(chan bool)

	go func() {
		<-kill
		close(kill)
		detachChan <- true
		close(detachChan)
		dead <- true
		close(dead)
	}()

	dummy := dummyVictim{kill, dead, detachChan}
	if !trySend(parent.attach(), victim(dummy), timeout) {
		return dummy, fmt.Errorf("failed to attach after %v", timeout)
	}
	return dummy, nil
}

func (dv dummyVictim) Kill() chan<- bool {
	return dv.kill
}

func (dv dummyVictim) Dead() <-chan bool {
	return dv.dead
}

func (dv dummyVictim) detach() <-chan bool {
	return dv.detachChan
}

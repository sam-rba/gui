package gui

import (
	"fmt"
	"testing"
)

// Kill the attachHandler with no victim attached.
func TestAttachHandlerKill(t *testing.T) {
	handler := newAttachHandler()
	if !trySend(handler.kill, true, timeout) {
		t.Errorf("kill attachHandler timed out after %v", timeout)
	}
	if _, ok := tryRecv(handler.dead, timeout); !ok {
		t.Errorf("no dead signal from attachHandler after %v", timeout)
	}
}

// Kill the attachHandler with a victim attached.
func TestAttachHandlerAttachKill(t *testing.T) {
	handler := newAttachHandler()
	victim, err := newDummyAttachable(handler)
	if err != nil {
		t.Error(err)
	}

	// Kill attachHandler.
	if !trySend(handler.kill, true, timeout) {
		t.Errorf("failed to kill attachHandler after %v", timeout)
	}
	if _, ok := tryRecv(handler.dead, timeout); !ok {
		t.Errorf("attachHandler not dead after %v", timeout)
	}

	// victim.Dead() should now be closed.
	if _, notClosed := <-victim.Dead(); notClosed {
		t.Errorf("victim not dead after killing attachHandler")
	}
}

// Detach the victim and attach another in its place.
func TestAttachHandlerReattach(t *testing.T) {
	handler := newAttachHandler()

	// Attach first victim.
	victim1, err := newDummyAttachable(handler)
	if err != nil {
		t.Error(err)
	}

	// Try to attach second victim while first still attached—should fail.
	if _, err := newDummyAttachable(handler); err == nil {
		t.Errorf("attachHandler accepted another victim while the first was still attached.")
	}

	// Detach first victim.
	if !trySend(victim1.Kill(), true, timeout) {
		t.Errorf("failed to kill first victim after %v", timeout)
	}
	if _, ok := tryRecv(victim1.Dead(), timeout); !ok {
		t.Errorf("first victim failed to die after %v", timeout)
	}

	// Attach second victim.
	if _, err := newDummyAttachable(handler); err != nil {
		t.Error(err)
	}

	handler.kill <- true
	<-handler.dead
}

type dummyAttachable struct {
	kill       chan<- bool
	dead       <-chan bool
	detachChan <-chan bool
}

// newDummyAttachable returns a dummyAttachable that is attached to parent,
// or error if the parent does not accept the attach.
func newDummyAttachable(parent killer) (attachable, error) {
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

	da := dummyAttachable{kill, dead, detachChan}
	if !trySend(parent.attach(), attachable(da), timeout) {
		return da, fmt.Errorf("failed to attach after %v", timeout)
	}
	return da, nil
}

func (da dummyAttachable) Kill() chan<- bool {
	return da.kill
}

func (da dummyAttachable) Dead() <-chan bool {
	return da.dead
}

func (da dummyAttachable) detach() <-chan bool {
	return da.detachChan
}

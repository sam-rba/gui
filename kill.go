package gui

// A Killable object can be told to shut down by sending a signal via the Kill() channel.
// As its last action, the object posts a signal to Dead() and closes both channels, indicating that it has finished shutting down.
type Killable interface {
	Kill() chan<- bool
	Dead() <-chan bool
}

type attachable interface {
	Killable
	// Sending to detach() will detach the object from the killer it is attached to.
	detach() <-chan bool
}

// A killer can kill the `victim' that is attached to it.
// The victim can attach itself to the killer by sending itself via the killer's attach() channel.
// The victim can detach itself by sending a signal via its own detach() channel.
//
// Only one victim can be attached to the killer at a time.
// Further messages sent on the attach() channel will block until the current victim is detached.
type killer interface {
	attach() chan<- attachable
}

// attachHandler implements killer. It allows victims to attach themselves via the attach channel.
// There can only be one attached victim at a time.
// If attachHandler is killed while a victim is attached, it kills the victim.
// When killed, the victim must detach itself before dying.
type attachHandler struct {
	attach chan<- attachable
	kill   chan<- bool
	dead   <-chan bool
}

func newAttachHandler() attachHandler {
	attach := make(chan attachable)
	kill := make(chan bool)
	dead := make(chan bool)

	go func() {
		defer func() {
			dead <- true
			close(dead)
		}()
		defer close(kill)
		defer close(attach)

		for {
			select {
			case victim := <-attach:
				select {
				case <-victim.detach():
				case <-kill:
					victim.Kill() <- true
					<-victim.detach()
					<-victim.Dead()
					return
				}
			case <-kill:
				return
			}
		}
	}()

	return attachHandler{attach, kill, dead}
}

package gui

// A Killable object can be told to shut down by sending a signal via the Kill() channel.
// As its last action, the object posts a signal to Dead() and closes both channels, indicating that it has finished shutting down.
type Killable interface {
	Kill() chan<- bool
	Dead() <-chan bool
}

// A killer can kill the object that is attached to it.
// The victim can attach itself to the killer by sending an attachMsg via the provided attach() channel.
// The attachMsg contains the victim itself and a `detach' channel.
// The victim can detatch itself from the killer by signalling over the `detach' channel.
//
// Only one victim can be attached to the killer at a time.
// Further messages sent on the attach() channel will block until the current victim is detached.
type killer interface {
	attach() chan<- attachMsg
}

// attachMsg is sent to a killer to attach the victim.
type attachMsg struct {
	victim Killable
	detach <-chan bool
}

// attachHandler implements killer. It allows victims to attach themselves via the attach channel.
// There can only be one attached victim at a time.
// If attachHandler is killed while a victim is attached, it kills the victim.
// When killed, the victim must detach itself before dying.
type attachHandler struct {
	attach chan<- attachMsg
	kill   chan<- bool
	dead   <-chan bool
}

func newAttachHandler() attachHandler {
	attach := make(chan attachMsg)
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
			var attached attachMsg

			select {
			case attached = <-attach:
			case <-kill:
				return
			}

		Attached:
			for {
				select {
				case <-attached.detach:
					break Attached
				case <-kill:
					attached.victim.Kill() <- true
					<-attached.detach
					<-attached.victim.Dead()
					return
				}
			}
		}
	}()

	return attachHandler{attach, kill, dead}
}

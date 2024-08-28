package gui

// A Killable object can be told to shut down by sending a signal via the Kill() channel.
// As its last action, the object posts a signal to Dead() and closes both channels, indicating that it has finished shutting down.
type Killable interface {
	Kill() chan<- bool
	Dead() <-chan bool
}

// A killer can kill the victim that is attached to it.
// The victim can attach itself to the killer by sending itself via the killer's attach() channel.
// The victim can detach itself by sending a signal via its own detach() channel.
//
// Only one victim can be attached to the killer at a time.
// Further messages sent on the attach() channel will block until the current victim is detached.
//
// If the killer is killed while a victim is attached, it kills the victim.
// When killed, the victim must detach itself before dying.
type killer interface {
	attach() chan<- victim

	Killable
}

type victim interface {
	// Sending to detach() will detach the victim from the killer it is attached to.
	detach() <-chan bool

	Killable
}

type _killer struct {
	attachChan chan<- victim
	kill       chan<- bool
	dead       <-chan bool
}

func newKiller() killer {
	attach := make(chan victim)
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

	return _killer{attach, kill, dead}
}

func (k _killer) attach() chan<- victim {
	return k.attachChan
}

func (k _killer) Kill() chan<- bool {
	return k.kill
}

func (k _killer) Dead() <-chan bool {
	return k.dead
}

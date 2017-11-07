package chatroom

import (
	"time"
)

type Event struct {
	Type      string // "join", "leave", or "message"
	Room			int
	User      string
	Timestamp int    // Unix timestamp (secs)
	Text      string // What the user said (if Type == "message")
}

type Subscription struct {
	Room		int
	Archive []Event      // All the events from the archive.
	New     <-chan Event // New events coming in.
}

type SubscribeCallback struct {
	Resp	chan Subscription
	Room int
}

type UnsubscribeCallback struct {
	SNew	<-chan Event
	Room int
}

// Owner of a subscription must cancel it when they stop listening to events.
func (s Subscription) Cancel(room int) {
	unsub := UnsubscribeCallback{s.New, room}
	unsubscribe <- unsub // Unsubscribe the channel.
	drain(s.New)         // Drain it, just in case there was a pending publish.
}

func newEvent(typ string, room int, user, msg string) Event {
	return Event{typ, room, user, int(time.Now().Unix()), msg}
}

func Subscribe(room int) Subscription {
	resp := make(chan Subscription)
	cb := SubscribeCallback{resp, room}
	subscribe <- cb
	return <-resp
}

func Join(room int, user string) {
	publish <- newEvent("join", room, user, "")
}

func Say(room int, user string, message string) {
	publish <- newEvent("message", room, user, message)
}

func Leave(room int, user string) {
	publish <- newEvent("leave", room, user, "")
}

const archiveSize = 10

var (
	// Send a channel here to get room events back.  It will send the entire
	// archive initially, and then new messages as they come in.
	subscribe = make(chan SubscribeCallback, 10)
	// Send a channel here to unsubscribe.
	unsubscribe = make(chan UnsubscribeCallback, 10)
	// Send events here to publish them.
	publish = make(chan Event, 10)
)

type Room struct{
	Archive	[]Event
	Subscribers []chan Event
}

// This function loops forever, handling the chat room pubsub
func chatroom() {
	capacity := 10
	rooms := make([]Room, 0, 100)
	for i:=0; i<100; i++ {
		archive := make([]Event, 0, capacity)
		subscribers := make([]chan Event, 0, capacity)
		rooms = append(rooms, Room{archive, subscribers})
	}

	for {
		select {
		case cb := <-subscribe:
			roomNo := cb.Room
			print("subscribe room:", roomNo, "\n")
			var events []Event
			for _, v := range rooms[roomNo].Archive {
				events = append(events, v)
			}
			subscriber := make(chan Event, 10)
			rooms[roomNo].Subscribers = append(rooms[roomNo].Subscribers, subscriber)
			cb.Resp <- Subscription{roomNo, events, subscriber}

		case event := <-publish:
			roomNo := event.Room
			print("publish room:", roomNo, "\n")
			for _, v := range rooms[roomNo].Subscribers {
				v <- event
			}
			if len(rooms[roomNo].Archive) >= archiveSize {
				rooms[roomNo].Archive = rooms[roomNo].Archive[1:]
			}
			rooms[roomNo].Archive = append(rooms[roomNo].Archive, event)

		case unsub := <-unsubscribe:
			roomNo := unsub.Room
			for i, v := range rooms[roomNo].Subscribers {
				if v == unsub.SNew {
					rooms[roomNo].Subscribers = append(rooms[roomNo].Subscribers[:i-1], rooms[roomNo].Subscribers[i:]...)
					break
				}
			}
		}
	}
}

func init() {
	go chatroom()
}

// Helpers

// Drains a given channel of any messages.
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

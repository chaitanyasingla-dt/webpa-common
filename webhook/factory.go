package webhook

import (
	"encoding/json"
	"github.com/Comcast/webpa-common/httperror"
	AWS "github.com/Comcast/webpa-common/webhook/aws"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

const (
	DEFAULT_UNDERTAKER_INTERVAL time.Duration = time.Minute
)

// Factory is a classic Factory Object for various webhook things.
type Factory struct {
	// Other configuration stuff can go here

	// Tick is an optional function that produces a channel for time ticks.
	// Test code can set this field to something that returns a channel under the control of the test.
	Tick func(time.Duration) <-chan time.Time `json:"-"`

	// UndertakerInterval is how often the undertaker is invoked
	UndertakerInterval time.Duration `json:"undertakerInterval"`

	// undertaker is set by clients after reading in a Factory from some external source.
	// The associated undertaker is immutable after construction.
	undertaker func([]W) []W `json:"-"`

	// internal handler for webhook
	m *monitor `json:"-"`

	// internal handler for AWS SNS Server
	AWS.Notifier `json:"-"`

	// StartConfig is the contains the data need to obtain the current system's listeners
	Start *StartConfig `json:"start"`
}

// NewFactory creates a Factory from a Viper environment.  This function always returns
// a non-nil Factory instance.
//
// This example uses Viper, which I highly recommend.  You could just pass an io.Reader too, and use
// the encoding/json package.  In any case, allowing the configuration source to be nil makes a lot
// of things easier on clients, like creating a test Factory for tests in client code.
func NewFactory(v *viper.Viper) (f *Factory, err error) {
	f = &Factory{
		/* put in any system defaults here.  they won't be overridden by Viper unless they are present in external configuration */
		UndertakerInterval: DEFAULT_UNDERTAKER_INTERVAL,
	}

	// allowing the viper instance to be nil allows a client to do
	// NewFactory(nil) to get a default Factory instance
	if v != nil {
		err = v.Unmarshal(f)
		if err != nil {
			return
		}
	}

	if v != nil {
		f.Start = NewStartFactory(v.Sub("start"))
	} else {
		f.Start = NewStartFactory(nil)
	}

	f.undertaker = f.Prune
	f.Notifier, err = AWS.NewNotifier(v)

	return
}

func (f *Factory) SetList(ul UpdatableList) {
	f.m.list = ul
}

func (f *Factory) Prune(items []W) (list []W) {
	for i := 0; i < len(items); i++ {
		if items[i].Until.After(time.Now()) {
			list = append(list, items[i])
		}
	}

	return
}

// NewRegistryAndHandler returns a List instance for accessing webhooks and an HTTP handler
// which can receive updates from external systems.
func (f *Factory) NewRegistryAndHandler() (Registry, http.Handler) {
	tick := f.Tick
	if tick == nil {
		tick = time.Tick
	}

	monitor := &monitor{
		list:             NewList(nil),
		undertaker:       f.undertaker,
		changes:          make(chan []W, 10),
		undertakerTicker: tick(f.UndertakerInterval),
	}
	f.m = monitor
	f.m.Notifier = f.Notifier

	reg := NewRegistry(f.m)

	go monitor.listen()
	return reg, monitor
}

func (f *Factory) SetExternalUpdate(fn func([]W)) {
	f.m.externalUpdate = fn
}

// monitor is an internal type that listens for webhook updates, invokes
// the undertaker at specified intervals, and responds to HTTP requests.
type monitor struct {
	list             UpdatableList
	undertaker       func([]W) []W
	changes          chan []W
	undertakerTicker <-chan time.Time
	AWS.Notifier
	externalUpdate   func([]W)
}

func (m *monitor) listen() {
	for {
		select {
		case update := <-m.changes:
			m.list.Update(update)
			
			if m.externalUpdate != nil {
				m.externalUpdate(update)
			}
		case <-m.undertakerTicker:
			m.list.Filter(m.undertaker)
		}
	}
}

func (m *monitor) sendNewHooks(newHooks []W) {
	select {
	case m.changes <- newHooks:
	default:
	}
}

// ServeHTTP is used as POST handler for AWS SNS
// It transforms the message containing webhook to []W and updates the webhook list
func (m *monitor) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// transform a request into a []byte
	message := m.NotificationHandle(response, request)
	if message == nil {
		return
	}

	// transform message to W
	var newHook W
	var newHooks []W
	var oldHook oldW
	var oldHooks []oldW
	if err := json.Unmarshal(message, &newHook); err == nil {
		m.sendNewHooks([]W{newHook})
	} else if err := json.Unmarshal(message, &newHooks); err == nil {
		m.sendNewHooks(newHooks)
	} else if err := json.Unmarshal(message, &oldHook); err == nil {
		newHook = doOldHookConvert(oldHook)
		m.sendNewHooks([]W{newHook})
	} else if err := json.Unmarshal(message, &oldHooks); err == nil {
		for _, oldHook := range oldHooks {
			newHooks = append(newHooks, doOldHookConvert(oldHook))
		}
		m.sendNewHooks(newHooks)
	} else {
		httperror.Format(response, http.StatusBadRequest, "Notification Message JSON unmarshall failed")
	}
}

package xhttp

import (
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/Comcast/webpa-common/logging"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewServerLogger(t *testing.T, logger log.Logger) {
	var (
		assert       = assert.New(t)
		require      = require.New(t)
		serverLogger = NewServerLogger(logger, "test")
	)

	require.NotNil(serverLogger)
	assert.NotPanics(func() {
		serverLogger.Println("this is a message")
	})
}

func TestNewServerLogger(t *testing.T) {
	t.Run("NilLogger", func(t *testing.T) {
		testNewServerLogger(t, nil)
	})

	t.Run("CustomLogger", func(t *testing.T) {
		testNewServerLogger(t, logging.NewTestLogger(nil, t))
	})
}

func testNewServerConnStateLogger(t *testing.T, logger log.Logger) {
	var (
		assert    = assert.New(t)
		require   = require.New(t)
		connState = NewServerConnStateLogger(logger, "test")
	)

	require.NotNil(connState)
	assert.NotPanics(func() {
		connState(new(net.IPConn), http.StateNew)
	})
}

func TestNewServerConnStateLogger(t *testing.T) {
	t.Run("NilLogger", func(t *testing.T) {
		testNewServerConnStateLogger(t, nil)
	})

	t.Run("CustomLogger", func(t *testing.T) {
		testNewServerConnStateLogger(t, logging.NewTestLogger(nil, t))
	})
}

const (
	expectedCertificateFile = "certificateFile"
	expectedKeyFile         = "keyFile"
)

// startOptions generates the various permutations of StartOptions that we test with.
// Each options struct can be further modified by tests.
func startOptions(t *testing.T) []StartOptions {
	var o []StartOptions

	for _, logger := range []log.Logger{nil, logging.NewTestLogger(nil, t)} {
		for _, disableKeepAlives := range []bool{false, true} {
			o = append(o, StartOptions{
				Logger:            logger,
				DisableKeepAlives: disableKeepAlives,
			})
		}
	}

	return o
}

func testNewStarterListenAndServe(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	for _, o := range startOptions(t) {
		t.Logf("StartOptions: %v", o)

		for _, expectedError := range []error{errors.New("expected"), http.ErrServerClosed} {
			httpServer := new(mockHTTPServer)

			httpServer.On("SetKeepAlivesEnabled", !o.DisableKeepAlives).Once()
			httpServer.On("ListenAndServe").Return(expectedError).Once()

			starter := NewStarter(o, httpServer)
			require.NotNil(starter)

			assert.NotPanics(func() {
				assert.Equal(expectedError, starter())
			})

			httpServer.AssertExpectations(t)
		}
	}
}

func testNewStarterServe(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	for _, o := range startOptions(t) {
		t.Logf("StartOptions: %v", o)

		for _, expectedError := range []error{errors.New("expected"), http.ErrServerClosed} {
			var (
				listener   = new(mockListener)
				httpServer = new(mockHTTPServer)
			)

			httpServer.On("SetKeepAlivesEnabled", !o.DisableKeepAlives).Once()
			httpServer.On("Serve", listener).Return(expectedError).Once()
			o.Listener = listener

			starter := NewStarter(o, httpServer)
			require.NotNil(starter)

			assert.NotPanics(func() {
				assert.Equal(expectedError, starter())
			})

			listener.AssertExpectations(t)
			httpServer.AssertExpectations(t)
		}
	}
}

func testNewStarterListenAndServeTLS(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	for _, o := range startOptions(t) {
		t.Logf("StartOptions: %v", o)

		for _, expectedError := range []error{errors.New("expected"), http.ErrServerClosed} {
			httpServer := new(mockHTTPServer)

			httpServer.On("SetKeepAlivesEnabled", !o.DisableKeepAlives).Once()
			httpServer.On("ListenAndServeTLS", expectedCertificateFile, expectedKeyFile).Return(expectedError).Once()
			o.CertificateFile = expectedCertificateFile
			o.KeyFile = expectedKeyFile

			starter := NewStarter(o, httpServer)
			require.NotNil(starter)

			assert.NotPanics(func() {
				assert.Equal(expectedError, starter())
			})

			httpServer.AssertExpectations(t)
		}
	}
}

func testNewStarterServeTLS(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	for _, o := range startOptions(t) {
		t.Logf("StartOptions: %v", o)

		for _, expectedError := range []error{errors.New("expected"), http.ErrServerClosed} {
			var (
				listener   = new(mockListener)
				httpServer = new(mockHTTPServer)
			)

			httpServer.On("SetKeepAlivesEnabled", !o.DisableKeepAlives).Once()
			httpServer.On("ServeTLS", listener, expectedCertificateFile, expectedKeyFile).Return(expectedError).Once()
			o.Listener = listener
			o.CertificateFile = expectedCertificateFile
			o.KeyFile = expectedKeyFile

			starter := NewStarter(o, httpServer)
			require.NotNil(starter)

			assert.NotPanics(func() {
				assert.Equal(expectedError, starter())
			})

			listener.AssertExpectations(t)
			httpServer.AssertExpectations(t)
		}
	}
}

func TestNewStarter(t *testing.T) {
	t.Run("ListenAndServe", testNewStarterListenAndServe)
	t.Run("Serve", testNewStarterServe)
	t.Run("ListenAndServeTLS", testNewStarterListenAndServeTLS)
	t.Run("ServeTLS", testNewStarterServeTLS)
}

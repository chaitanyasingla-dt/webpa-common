package conveymetric

import (
	"github.com/Comcast/webpa-common/convey"
	"github.com/Comcast/webpa-common/xmetrics"
	"github.com/Comcast/webpa-common/xmetrics/xmetricstest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConveyMetric(t *testing.T) {
	assert := assert.New(t)

	//namespace, subsystem, name := "test", "basic", "hardware"

	gauge := xmetricstest.NewGauge("hardware")

	conveyMetric := NewConveyMetric(gauge, "hw-model", "model")

	dec, err := conveyMetric.Update(convey.C{"data": "neat", "hw-model": "hardware123abc"})
	assert.NoError(err)
	assert.Equal(float64(1), gauge.With("model", "hardware123abc").(xmetrics.Valuer).Value())
	// remove the update
	dec()

	assert.Equal(float64(0), gauge.With("model", "hardware123abc").(xmetrics.Valuer).Value())

	// try with no `hw_model`
	dec, err = conveyMetric.Update(convey.C{"data": "neat"})
	assert.Equal(float64(1), gauge.With("model", UnknownLabel).(xmetrics.Valuer).Value())

	// remove the update
	dec()
	assert.Equal(float64(0), gauge.With("model", UnknownLabel).(xmetrics.Valuer).Value())
}

func TestGetValidKeyName(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("basic", getValidKeyName("basic"))
	assert.Equal(UnknownLabel, getValidKeyName(""))
	assert.Equal(UnknownLabel, getValidKeyName(" "))
	assert.Equal("hw_model", getValidKeyName("hw-model"))
	assert.Equal("hw_model", getValidKeyName("hw model"))
	assert.Equal("hw_model", getValidKeyName("hw	model"))
	assert.Equal("hw_model", getValidKeyName(" hw	model"))
	assert.Equal("hw_model", getValidKeyName(" hw	model@"))
	assert.Equal("hw_model", getValidKeyName("hw@model"))
	assert.Equal("hw_model", getValidKeyName("hw@ model"))
}

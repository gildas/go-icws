package icws_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gildas/go-icws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanMarshalTime(t *testing.T) {
	expected := []byte(`{"Time":"20210706T182219Z"}`)
	var icwsTime = icws.Time(time.Date(2021, 07,06, 18, 22, 19, 0, time.UTC))

	payload, err := json.Marshal(struct{Time icws.Time}{Time: icwsTime})
	require.Nil(t, err)
	assert.Equal(t, expected, payload)
}

func TestCanUnmarshalTime(t *testing.T) {
	payload := []byte(`{"Time":"20210706T182219Z"}`)

	data := struct{Time icws.Time}{}

	err := json.Unmarshal(payload, &data)
	require.Nil(t, err)
	assert.Equal(t, 2021, time.Time(data.Time).Year())
	assert.Equal(t, time.July, time.Time(data.Time).Month())
	assert.Equal(t, 6, time.Time(data.Time).Day())
}

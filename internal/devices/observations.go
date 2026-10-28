package devices

// Derived synthetic inputs: no source audio or screen pixels travel on the bus.
type AudioFeatures struct {
	RMS      float64    `json:"rms"`
	Spectrum [3]float64 `json:"spectrum"`
	Beat     bool       `json:"beat"`
	BPM      float64    `json:"bpm"`
	Quality  string     `json:"quality"`
}
type Presence struct {
	State   string `json:"state"`
	Quality string `json:"quality"`
}
type ScreenColors struct {
	Epoch    string `json:"epoch"`
	Sequence uint64 `json:"sequence"`
	Colors   []RGB  `json:"colors"`
	Error    string `json:"error,omitempty"`
	Quality  string `json:"quality"`
}
type MQTTObservation struct {
	Phase    string  `json:"phase"`
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Quality  string  `json:"quality"`
}

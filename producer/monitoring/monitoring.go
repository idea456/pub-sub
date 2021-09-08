package monitoring

import "github.com/prometheus/client_golang/prometheus"

// Don't forget to register the metrics here also!

func Init() error {
	if err := prometheus.Register(TotalRequests); err != nil {
		return err
	}
	if err := prometheus.Register(ResponseStatus); err != nil {
		return err
	}
	if err := prometheus.Register(Latency); err != nil {
		return err
	}
	return nil
}

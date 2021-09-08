package monitoring

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		method := r.Method

		// timer := prometheus.NewTimer(latency.WithLabelValues(path, method))

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		ResponseStatus.WithLabelValues(strconv.Itoa(rw.statusCode), path, method).Inc()
		TotalRequests.WithLabelValues(path, method).Inc()
		// timer.ObserveDuration()
	})
}

func ObserveResponse(corrId string, msgs <-chan amqp.Delivery) {
	go func() {
		for msg := range msgs {
			if corrId == msg.CorrelationId {
				log.Printf("Message has been received by server : %s", msg.Body)
				PrintTimer.ObserveDuration()
			}
		}
		log.Printf("Doneeee")
	}()
}

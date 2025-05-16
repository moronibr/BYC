package monitoring

import (
	"testing"
	"time"
)

type alertCatcher struct {
	alerts []*Alert
}

func (a *alertCatcher) OnMetricUpdate(metric Metric) {
	// No-op for this test
}

func catchAlerts(m *Monitor) chan *Alert {
	ch := make(chan *Alert, 10)
	go func() {
		for {
			select {
			case alert := <-m.alertChan:
				ch <- alert
			case <-time.After(100 * time.Millisecond):
				return
			}
		}
	}()
	return ch
}

func TestCheckAlerts_CPU(t *testing.T) {
	m := NewMonitor(nil, nil, time.Second)
	ch := catchAlerts(m)
	m.checkAlerts(Metric{Type: MetricCPUUsage, Value: 95, Timestamp: time.Now()})
	alert := <-ch
	if alert.Level != AlertCritical {
		t.Errorf("expected critical alert, got %v", alert.Level)
	}
	m.checkAlerts(Metric{Type: MetricCPUUsage, Value: 80, Timestamp: time.Now()})
	alert = <-ch
	if alert.Level != AlertWarning {
		t.Errorf("expected warning alert, got %v", alert.Level)
	}
}

func TestCheckAlerts_Memory(t *testing.T) {
	m := NewMonitor(nil, nil, time.Second)
	ch := catchAlerts(m)
	m.checkAlerts(Metric{Type: MetricMemoryUsage, Value: 85, Timestamp: time.Now()})
	alert := <-ch
	if alert.Level != AlertCritical {
		t.Errorf("expected critical alert, got %v", alert.Level)
	}
	m.checkAlerts(Metric{Type: MetricMemoryUsage, Value: 65, Timestamp: time.Now()})
	alert = <-ch
	if alert.Level != AlertWarning {
		t.Errorf("expected warning alert, got %v", alert.Level)
	}
}

func TestCheckAlerts_Disk(t *testing.T) {
	m := NewMonitor(nil, nil, time.Second)
	ch := catchAlerts(m)
	m.checkAlerts(Metric{Type: MetricDiskUsage, Value: 95, Timestamp: time.Now()})
	alert := <-ch
	if alert.Level != AlertCritical {
		t.Errorf("expected critical alert, got %v", alert.Level)
	}
	m.checkAlerts(Metric{Type: MetricDiskUsage, Value: 80, Timestamp: time.Now()})
	alert = <-ch
	if alert.Level != AlertWarning {
		t.Errorf("expected warning alert, got %v", alert.Level)
	}
}

func TestCheckAlerts_BlockTime(t *testing.T) {
	m := NewMonitor(nil, nil, time.Second)
	ch := catchAlerts(m)
	m.checkAlerts(Metric{Type: MetricBlockTime, Value: 35, Timestamp: time.Now()})
	alert := <-ch
	if alert.Level != AlertWarning {
		t.Errorf("expected warning alert, got %v", alert.Level)
	}
}

func TestCheckAlerts_MempoolSize(t *testing.T) {
	m := NewMonitor(nil, nil, time.Second)
	ch := catchAlerts(m)
	m.checkAlerts(Metric{Type: MetricMempoolSize, Value: 20000, Timestamp: time.Now()})
	alert := <-ch
	if alert.Level != AlertWarning {
		t.Errorf("expected warning alert, got %v", alert.Level)
	}
}

func TestCheckAlerts_NetworkPeers(t *testing.T) {
	m := NewMonitor(nil, nil, time.Second)
	ch := catchAlerts(m)
	m.checkAlerts(Metric{Type: MetricNetworkPeers, Value: 2, Timestamp: time.Now()})
	alert := <-ch
	if alert.Level != AlertWarning {
		t.Errorf("expected warning alert, got %v", alert.Level)
	}
}

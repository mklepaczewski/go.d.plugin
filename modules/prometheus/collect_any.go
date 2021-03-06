package prometheus

import (
	"github.com/netdata/go.d.plugin/pkg/prometheus"
)

func (p *Prometheus) collectAny(mx map[string]int64, pms prometheus.Metrics, meta prometheus.Metadata) {
	name := pms[0].Name()
	if !p.cache.has(name) {
		s, err := newAnySplit(pms)
		if err != nil {
			p.skipMetrics[name] = true
			p.Infof("skip metric '%s': %v", name, err)
			return
		}
		p.cache.put(name, &cacheEntry{
			split:  s,
			charts: make(chartsCache),
			dims:   make(dimsCache),
		})
	}

	cache := p.cache.get(name)

	for _, pm := range pms {
		chartID := cache.split.chartID(pm)
		dimID := cache.split.dimID(pm)
		dimName := cache.split.dimName(pm)

		mx[dimID] = int64(pm.Value * precision)

		if !cache.hasChart(chartID) {
			chart := anyChart(chartID, pm, meta)
			cache.putChart(chartID, chart)
			if err := p.Charts().Add(chart); err != nil {
				p.Warning(err)
			}
		}
		if !cache.hasDim(dimID) {
			cache.putDim(dimID)
			chart := cache.getChart(chartID)
			dim := anyChartDimension(dimID, dimName, pm, meta)
			if err := chart.AddDim(dim); err != nil {
				p.Warning(err)
			}
			chart.MarkNotCreated()
		}
	}

	var reSplit bool
	for _, chart := range cache.charts {
		if len(chart.Dims) > maxDim {
			reSplit = true
			break
		}
	}
	if reSplit {
		p.cleanupAnyMetric(name)
		p.collectAny(mx, pms, meta)
	}
}

func (p *Prometheus) cleanupAnyMetric(name string) {
	if !p.cache.has(name) {
		return
	}
	defer p.cache.remove(name)

	for _, chart := range p.cache.get(name).charts {
		for _, dim := range chart.Dims {
			_ = chart.MarkDimRemove(dim.ID, true)
		}
		chart.MarkRemove()
		chart.MarkNotCreated()
	}
}

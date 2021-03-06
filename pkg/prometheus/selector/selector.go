package selector

import (
	"github.com/netdata/go.d.plugin/pkg/matcher"

	"github.com/prometheus/prometheus/pkg/labels"
)

type Selector interface {
	Matches(lbs labels.Labels) bool
}

const (
	OpEqual             = "="
	OpNegEqual          = "!="
	OpRegexp            = "=~"
	OpNegRegexp         = "!~"
	OpSimplePatterns    = "=*"
	OpNegSimplePatterns = "!*"
)

type labelSelector struct {
	name string
	m    matcher.Matcher
}

func (s labelSelector) Matches(lbs labels.Labels) bool {
	if s.name == labels.MetricName {
		return s.m.MatchString(lbs[0].Value)
	}
	if label, ok := lookupLabel(s.name, lbs[1:]); ok {
		return s.m.MatchString(label.Value)
	}
	return false
}

func lookupLabel(name string, lbs labels.Labels) (labels.Label, bool) {
	for _, label := range lbs {
		if label.Name == name {
			return label, true
		}
	}
	return labels.Label{}, false
}

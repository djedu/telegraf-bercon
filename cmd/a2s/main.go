//go:generate ../../../tools/readme_config_includer/generator
package a2s

import (
	_ "embed"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/woozymasta/bercon-cli/pkg/bercon"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type A2S struct {
	Servers [][]string `toml:"servers"`
}

func (*A2S) SampleConfig() string {
	return sampleConfig
}

func (s *A2S) Init() error {
	for _, server := range s.Servers {
		if len(server) != 2 {
			return errors.New("incorrect server config")
		}
	}
	return nil
}

func (s *A2S) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range s.Servers {
		wg.Add(1)
		go func(addr, passwd string) {
			defer wg.Done()

			// Open RCON connection
			client, err := bercon.Open(addr, passwd)
			if err != nil {
				acc.AddError(err)
				return
			}
			defer client.Close()

			// // Connect and send the request
			// client, err := rcon.Dial(addr, passwd)
			// if err != nil {
			// 	acc.AddError(err)
			// 	return
			// }
			// defer client.Close()

			t := time.Now()
			response, err := client.Send("players")
			if err != nil {
				acc.AddError(err)
				return
			}

			// Generate the metric and add it to the accumulator
			m, err := parseResponse(addr, string(response), t)
			if err != nil {
				acc.AddError(err)
				return
			}
			acc.AddMetric(m)
		}(server[0], server[1])
	}

	wg.Wait()
	return nil
}

func parseResponse(addr, response string, t time.Time) (telegraf.Metric, error) {
	rows := strings.Split(response, "\n")
	if len(rows) < 2 {
		return nil, errors.New("bad response")
	}

	// Parse the columns
	columns := strings.Fields(rows[1])
	if len(columns) != 10 {
		return nil, errors.New("not enough columns")
	}

	cpu, err := strconv.ParseFloat(columns[0], 32)
	if err != nil {
		return nil, err
	}
	netIn, err := strconv.ParseFloat(columns[1], 64)
	if err != nil {
		return nil, err
	}
	netOut, err := strconv.ParseFloat(columns[2], 64)
	if err != nil {
		return nil, err
	}
	uptimeMinutes, err := strconv.ParseFloat(columns[3], 64)
	if err != nil {
		return nil, err
	}
	maps, err := strconv.ParseFloat(columns[4], 64)
	if err != nil {
		return nil, err
	}
	fps, err := strconv.ParseFloat(columns[5], 64)
	if err != nil {
		return nil, err
	}
	players, err := strconv.ParseFloat(columns[6], 64)
	if err != nil {
		return nil, err
	}
	svms, err := strconv.ParseFloat(columns[7], 64)
	if err != nil {
		return nil, err
	}
	msVar, err := strconv.ParseFloat(columns[8], 64)
	if err != nil {
		return nil, err
	}
	tick, err := strconv.ParseFloat(columns[9], 64)
	if err != nil {
		return nil, err
	}

	// Construct the metric
	tags := map[string]string{"host": addr}
	fields := map[string]interface{}{
		"cpu":            cpu,
		"net_in":         netIn,
		"net_out":        netOut,
		"uptime_minutes": uptimeMinutes,
		"maps":           maps,
		"fps":            fps,
		"players":        players,
		"sv_ms":          svms,
		"variance_ms":    msVar,
		"tick_ms":        tick,
	}
	return metric.New("a2s", tags, fields, t, telegraf.Gauge), nil
}

func init() {
	inputs.Add("a2s", func() telegraf.Input {
		return &A2S{}
	})
}

package simulator

import (
	"context"
	"runtime/debug"
	"strconv"

	"github.com/simimpact/srsim/pkg/gcs"
	"github.com/simimpact/srsim/pkg/model"
	"github.com/simimpact/srsim/pkg/simulator/workerpool"
	"github.com/simimpact/srsim/pkg/statistics/agg"
	"google.golang.org/protobuf/proto"
)

var (
	sha1ver   string
	buildTime string
	modified  bool
)

func init() {
	info, _ := debug.ReadBuildInfo()
	for _, bs := range info.Settings {
		if bs.Key == "vcs.revision" {
			sha1ver = bs.Value
		}
		if bs.Key == "vcs.time" {
			buildTime = bs.Value
		}
		if bs.Key == "vcs.modified" {
			bv, _ := strconv.ParseBool(bs.Value)
			modified = bv
		}
	}
}

func Version() string {
	return sha1ver
}

func Run(ctx context.Context, list *gcs.ActionList, cfg *model.SimConfig) (*model.SimulationResult, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var aggregators []agg.Aggregator
	for _, aggregator := range agg.Aggregators() {
		a, err := aggregator(cfg)
		if err != nil {
			return nil, err
		}
		aggregators = append(aggregators, a)
	}

	resp := make(chan *model.IterationResult)
	errChan := make(chan error)
	pool := workerpool.New(
		ctx,
		int(cfg.WorkerCount),
		resp,
		errChan,
	)

	go func() {
		for i := 0; i < int(cfg.Iterations); i++ {
			j := proto.Clone(cfg).(*model.SimConfig)
			err := pool.QueueJob(workerpool.Job{
				Script: list,
				Config: j,
			})
			if err != nil {
				//context must have been cancelled
				return
			}
		}
	}()

	//get results back
	for i := 0; i < int(cfg.Iterations); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errChan:
			return nil, err
		case result := <-resp:
			for _, a := range aggregators {
				a.Add(result)
			}
		}
	}

	//stats aggregation should happen here and make us a result?
	result := &model.SimulationResult{
		SimVersion: &sha1ver,
		Modified:   &modified,
		BuildDate:  buildTime,
	}

	return result, nil
}

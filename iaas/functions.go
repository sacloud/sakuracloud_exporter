package iaas

import (
	"context"
	"sync"
)

type perZoneQueryFunc func(ctx context.Context, zone string) ([]interface{}, error)

func queryToZones(ctx context.Context, zones []string, query perZoneQueryFunc) ([]interface{}, error) {
	var wg sync.WaitGroup
	wg.Add(len(zones))

	type result struct {
		results []interface{}
		err     error
	}

	resCh := make(chan *result)
	defer close(resCh)

	for _, zone := range zones {
		go func(zone string) {
			res, err := query(ctx, zone)
			resCh <- &result{
				results: res,
				err:     err,
			}
		}(zone)
	}

	var results []interface{}
	var err error
	go func() {
		for res := range resCh {
			if err == nil {
				if res.err != nil {
					err = res.err
				} else {
					results = append(results, res.results...)
				}
			}
			wg.Done()
		}
	}()

	wg.Wait()
	return results, err
}

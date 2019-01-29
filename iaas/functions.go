package iaas

import (
	"sync"

	sakuraAPI "github.com/sacloud/libsacloud/api"
)

func queryToZones(client *sakuraAPI.Client, zones []string, query func(*sakuraAPI.Client) ([]interface{}, error)) ([]interface{}, error) {
	var wg sync.WaitGroup
	wg.Add(len(zones))

	resCh := make(chan []interface{})
	errCh := make(chan error)

	for _, zone := range zones {
		c := client.Clone()
		c.Zone = zone

		go func(c *sakuraAPI.Client) {
			res, err := query(c)
			if err != nil {
				errCh <- err
				return
			}
			resCh <- res
		}(c)
	}

	var results []interface{}
	var err error
	go func() {
		for {
			if err != nil {
				wg.Done()
				return
			}
			select {
			case res := <-resCh:
				results = append(results, res...)
			case e := <-errCh:
				err = e
			}
			wg.Done()
		}
	}()

	wg.Wait()
	return results, err
}

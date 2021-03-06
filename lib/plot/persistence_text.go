package plot

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/uol/gobol"
)

func (persist *persistence) GetTST(
	keyspace,
	key string,
	start,
	end int64,
	tuuid bool,
	search *regexp.Regexp,
) (TextPnts, int, gobol.Error) {

	if tuuid {
		return persist.getTSTuuid(keyspace, key, start, end, search)
	}

	return persist.getTSTstamp(keyspace, key, start, end, search)
}

func (persist *persistence) getTSTuuid(
	keyspace,
	key string,
	start,
	end int64,
	search *regexp.Regexp,
) ([]TextPnt, int, gobol.Error) {
	track := time.Now()
	start--
	end++

	var date int64
	var value string
	var err error

	for _, cons := range persist.consistencies {
		iter := persist.cassandra.Query(
			fmt.Sprintf(
				`SELECT toUnixTimestamp(date), value FROM %v.ts_text WHERE id= ? AND date > maxTimeuuid(?) AND date < minTimeuuid(?) ALLOW FILTERING`,
				keyspace,
			),
			key,
			start,
			end,
		).Consistency(cons).RoutingKey([]byte(key)).Iter()

		points := []TextPnt{}
		var count int

		for iter.Scan(&date, &value) {

			add := true

			if search != nil && !search.MatchString(value) {
				add = false
			}

			if add {
				count++
				point := TextPnt{
					Date:  date,
					Value: value,
				}
				points = append(points, point)
			}
		}

		if err = iter.Close(); err != nil {

			gblog.WithFields(logrus.Fields{
				"package": "plot/persistence",
				"func":    "getTSTuuid",
			}).Error(err)

			if err == gocql.ErrNotFound {
				return []TextPnt{}, 0, errNoContent("getTSTuuid")
			}

			statsSelectQerror(keyspace, "ts_text")
			continue
		}
		statsSelect(keyspace, "ts_text", time.Since(track))
		return points, count, nil
	}
	statsSelectFerror(keyspace, "ts_text")
	return []TextPnt{}, 0, errPersist("getTSTuuid", err)
}

func (persist *persistence) getTSTstamp(
	keyspace,
	key string,
	start,
	end int64,
	search *regexp.Regexp,
) ([]TextPnt, int, gobol.Error) {
	track := time.Now()
	start--
	end++

	var date int64
	var value string
	var err error

	for _, cons := range persist.consistencies {
		iter := persist.cassandra.Query(
			fmt.Sprintf(
				`SELECT date, value FROM %v.ts_text_stamp WHERE id= ? AND date > ? AND date < ? ALLOW FILTERING`,
				keyspace,
			),
			key,
			start,
			end,
		).Consistency(cons).RoutingKey([]byte(key)).Iter()

		points := []TextPnt{}
		var count int

		for iter.Scan(&date, &value) {
			add := true

			if search != nil && !search.MatchString(value) {
				add = false
			}

			if add {
				count++
				point := TextPnt{
					Date:  date,
					Value: value,
				}
				points = append(points, point)
			}
		}

		if err = iter.Close(); err != nil {

			gblog.WithFields(logrus.Fields{
				"package": "plot/persistence",
				"func":    "getTSTstamp",
			}).Error(err)

			if err == gocql.ErrNotFound {
				return []TextPnt{}, 0, errNoContent("getTSTstamp")
			}

			statsSelectQerror(keyspace, "ts_text_stamp")
			continue
		}
		statsSelect(keyspace, "ts_text_stamp", time.Since(track))
		return points, count, nil
	}
	statsSelectFerror(keyspace, "ts_text_stamp")
	return []TextPnt{}, 0, errPersist("getTSTstamp", err)
}

func (persist *persistence) fuseText(countF, countS int, first, second []TextPnt) []TextPnt {

	fused := make(TextPnts, countF+countS)
	var i, j, k int

	for i < countF && j < countS {
		if first[i].Date <= second[j].Date {
			fused[k] = first[i]
			i++
		} else {
			fused[k] = second[j]
			j++
		}
		k++
	}
	if i < countF {
		for p := i; p < countF; p++ {
			fused[k] = first[p]
			k++
		}
	} else {
		for p := j; p < countS; p++ {
			fused[k] = second[p]
			k++
		}
	}

	return fused
}

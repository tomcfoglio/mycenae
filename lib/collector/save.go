package collector

import (
	"fmt"
	"time"

	"github.com/uol/gobol"
)

func (collector *Collector) saveValue(packet Point) gobol.Error {
	return collector.persist.InsertPoint(
		packet.KsID,
		fmt.Sprintf("%v%v", packet.Bucket, packet.ID),
		packet.Timestamp,
		*packet.Message.Value,
	)
}

func (collector *Collector) saveTUUIDvalue(packet Point) gobol.Error {
	return collector.persist.InsertTUUIDpoint(
		packet.KsID,
		fmt.Sprintf("%v%v", packet.Bucket, packet.ID),
		packet.TimeUUID,
		*packet.Message.Value,
	)
}

func (collector *Collector) saveText(packet Point) gobol.Error {
	return collector.persist.InsertText(
		packet.KsID,
		fmt.Sprintf("%v%v", packet.Bucket, packet.ID),
		packet.Timestamp,
		packet.Message.Text,
	)
}

func (collector *Collector) saveTUUIDtext(packet Point) gobol.Error {
	return collector.persist.InsertTUUIDtext(
		packet.KsID,
		fmt.Sprintf("%v%v", packet.Bucket, packet.ID),
		packet.TimeUUID,
		packet.Message.Text,
	)
}

func (collector *Collector) saveError(
	recvTags map[string]string,
	metric,
	keyspace,
	esIndex,
	id,
	msg,
	errMsg string,
) gobol.Error {

	now := time.Now()
	ks := keyspace
	if keyspace == collector.settings.Cassandra.Keyspace {
		ks = "default"
	}
	statsUDPerror(ks, "number")

	idks := fmt.Sprintf("%s%s", id, keyspace)

	gerr := collector.persist.InsertError(idks, msg, errMsg, now)
	if gerr != nil {
		return gerr
	}

	var tags []Tag

	for k, v := range recvTags {
		tag := Tag{
			Key:   k,
			Value: v,
		}
		tags = append(tags, tag)
	}

	doc := StructV2Error{
		Key:    id,
		Metric: metric,
		Tags:   tags,
	}

	collector.persist.SendErrorToES(esIndex, "errortag", id, doc)

	return nil
}

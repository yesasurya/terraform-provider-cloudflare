package cache

import (
	"context"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CloudflareDnsRecordCache struct {
	ready   sync.Map
	records sync.Map
}

var dnsCache = CloudflareDnsRecordCache{}

func GetDnsRecordFromCache(ctx context.Context, d *schema.ResourceData, meta interface{}) (cloudflare.DNSRecord, error) {
	zoneId := d.Get("zone_id").(string)

	if _, exist := dnsCache.records.Load(zoneId); !exist {
		dnsCache.records.Store(zoneId, make(map[string]cloudflare.DNSRecord))
		client := meta.(*cloudflare.API)
		records, _ := client.DNSRecords(ctx, zoneId, cloudflare.DNSRecord{})
		for _, record := range records {
			zoneRecords, _ := dnsCache.records.Load(zoneId)
			zoneRecords.(map[string]cloudflare.DNSRecord)[record.ID] = record
		}
		dnsCache.ready.Store(zoneId, true)
	}

	// Sleep-wait
	for {
		if _, ready := dnsCache.ready.Load(zoneId); ready {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	zoneRecords, _ := dnsCache.records.Load(zoneId)
	return zoneRecords.(map[string]cloudflare.DNSRecord)[d.Id()], nil
}

package azurerm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/dns/mgmt/2018-03-01-preview/dns"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmDnsAAAARecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmDnsAaaaRecordCreateOrUpdate,
		Read:   resourceArmDnsAaaaRecordRead,
		Update: resourceArmDnsAaaaRecordCreateOrUpdate,
		Delete: resourceArmDnsAaaaRecordDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 10),
			Update: schema.DefaultTimeout(time.Minute * 10),
			Delete: schema.DefaultTimeout(time.Minute * 10),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"zone_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"records": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmDnsAaaaRecordCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).dnsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	zoneName := d.Get("zone_name").(string)

	if d.IsNewResource() {
		// first check if there's one in this subscription requiring import
		resp, err := client.Get(ctx, resGroup, zoneName, name, dns.AAAA)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Error checking for the existence of DNS AAAA Record %q (Zone %q / Resource Group %q): %+v", name, zoneName, resGroup, err)
			}
		}

		if resp.ID != nil {
			return tf.ImportAsExistsError("azurerm_dns_aaaa_record", *resp.ID)
		}
	}

	ttl := int64(d.Get("ttl").(int))
	tags := d.Get("tags").(map[string]interface{})

	records, err := expandAzureRmDnsAaaaRecords(d)
	if err != nil {
		return err
	}

	parameters := dns.RecordSet{
		Name: &name,
		RecordSetProperties: &dns.RecordSetProperties{
			Metadata:    expandTags(tags),
			TTL:         &ttl,
			AaaaRecords: &records,
		},
	}

	eTag := ""
	ifNoneMatch := "" // set to empty to allow updates to records after creation
	waitCtx, cancel := context.WithTimeout(ctx, d.Timeout(tf.TimeoutForCreateUpdate(d)))
	defer cancel()
	resp, err := client.CreateOrUpdate(waitCtx, resGroup, zoneName, name, dns.AAAA, parameters, eTag, ifNoneMatch)
	if err != nil {
		return err
	}

	if resp.ID == nil {
		return fmt.Errorf("Cannot read DNS AAAA Record %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*resp.ID)

	return resourceArmDnsAaaaRecordRead(d, meta)
}

func resourceArmDnsAaaaRecordRead(d *schema.ResourceData, meta interface{}) error {
	dnsClient := meta.(*ArmClient).dnsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	name := id.Path["AAAA"]
	zoneName := id.Path["dnszones"]

	resp, err := dnsClient.Get(ctx, resGroup, zoneName, name, dns.AAAA)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading DNS AAAA record %s: %v", name, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resGroup)
	d.Set("zone_name", zoneName)
	d.Set("ttl", resp.TTL)

	if err := d.Set("records", flattenAzureRmDnsAaaaRecords(resp.AaaaRecords)); err != nil {
		return err
	}
	flattenAndSetTags(d, resp.Metadata)

	return nil
}

func resourceArmDnsAaaaRecordDelete(d *schema.ResourceData, meta interface{}) error {
	dnsClient := meta.(*ArmClient).dnsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	name := id.Path["AAAA"]
	zoneName := id.Path["dnszones"]

	waitCtx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()
	resp, error := dnsClient.Delete(waitCtx, resGroup, zoneName, name, dns.AAAA, "")
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error deleting DNS AAAA Record %s: %+v", name, error)
	}

	return nil
}

func flattenAzureRmDnsAaaaRecords(records *[]dns.AaaaRecord) []string {
	results := make([]string, 0)

	if records != nil {
		for _, record := range *records {
			if record.Ipv6Address != nil {
				results = append(results, *record.Ipv6Address)
			}
		}
	}

	return results
}

func expandAzureRmDnsAaaaRecords(d *schema.ResourceData) ([]dns.AaaaRecord, error) {
	recordStrings := d.Get("records").(*schema.Set).List()
	records := make([]dns.AaaaRecord, 0)

	for _, v := range recordStrings {
		ipv6 := v.(string)
		record := dns.AaaaRecord{
			Ipv6Address: &ipv6,
		}
		records = append(records, record)
	}

	return records, nil
}

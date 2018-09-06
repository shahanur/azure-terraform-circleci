package azurerm

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/scheduler/mgmt/2016-03-01/scheduler"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmSchedulerJobCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSchedulerJobCollectionCreateUpdate,
		Read:   resourceArmSchedulerJobCollectionRead,
		Update: resourceArmSchedulerJobCollectionCreateUpdate,
		Delete: resourceArmSchedulerJobCollectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 30),
			Update: schema.DefaultTimeout(time.Minute * 30),
			Delete: schema.DefaultTimeout(time.Minute * 30),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z][-_a-zA-Z0-9]{0,99}$"),
					"Job Collection Name name must be 1 - 100 characters long, start with a letter and contain only letters, numbers, hyphens and underscores.",
				),
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"tags": tagsSchema(),

			"sku": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				ValidateFunc: validation.StringInSlice([]string{
					string(scheduler.Free),
					string(scheduler.Standard),
					string(scheduler.P10Premium),
					string(scheduler.P20Premium),
				}, true),
			},

			//optional
			"state": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(scheduler.Enabled),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				ValidateFunc: validation.StringInSlice([]string{
					string(scheduler.Enabled),
					string(scheduler.Suspended),
					string(scheduler.Disabled),
				}, true),
			},

			"quota": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						//max_job_occurrence doesn't seem to do anything and always remains empty

						"max_job_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"max_recurrence_frequency": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
							ValidateFunc: validation.StringInSlice([]string{
								string(scheduler.Minute),
								string(scheduler.Hour),
								string(scheduler.Day),
								string(scheduler.Week),
								string(scheduler.Month),
							}, true),
						},

						// API documentation states the MaxRecurrence.Interval "Gets or sets the interval between retries."
						// however it does appear it is the max interval allowed for recurrences
						"max_retry_interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Deprecated:   "Renamed to `max_recurrence_interval` to match azure",
							ValidateFunc: validation.IntAtLeast(1), //changes depending on the frequency, unknown maximums
						},

						"max_recurrence_interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1), //changes depending on the frequency, unknown maximums
						},
					},
				},
			},
		},
	}
}

func resourceArmSchedulerJobCollectionCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).schedulerJobCollectionsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	if d.IsNewResource() {
		// first check if there's one in this subscription requiring import
		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Error checking for the existence of Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
			}
		}

		if resp.ID != nil {
			return tf.ImportAsExistsError("azurerm_scheduler_job_collection", *resp.ID)
		}
	}

	location := azureRMNormalizeLocation(d.Get("location").(string))
	tags := d.Get("tags").(map[string]interface{})

	log.Printf("[DEBUG] Creating/updating Scheduler Job Collection %q (resource group %q)", name, resourceGroup)

	collection := scheduler.JobCollectionDefinition{
		Location: utils.String(location),
		Tags:     expandTags(tags),
		Properties: &scheduler.JobCollectionProperties{
			Sku: &scheduler.Sku{
				Name: scheduler.SkuDefinition(d.Get("sku").(string)),
			},
		},
	}

	if state, ok := d.Get("state").(string); ok {
		collection.Properties.State = scheduler.JobCollectionState(state)
	}
	collection.Properties.Quota = expandAzureArmSchedulerJobCollectionQuota(d)

	//create job collection
	waitCtx, cancel := context.WithTimeout(ctx, d.Timeout(tf.TimeoutForCreateUpdate(d)))
	defer cancel()
	collection, err := client.CreateOrUpdate(waitCtx, resourceGroup, name, collection)
	if err != nil {
		return fmt.Errorf("Error creating/updating Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	//ensure collection actually exists and we have the correct ID
	collection, err = client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("Error reading Scheduler Job Collection %q after create/update (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(*collection.ID)

	return resourceArmSchedulerJobCollectionRead(d, meta)
}

func resourceArmSchedulerJobCollectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).schedulerJobCollectionsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["jobCollections"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Reading Scheduler Job Collection %q (resource group %q)", name, resourceGroup)

	collection, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(collection.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	//standard properties
	d.Set("name", collection.Name)
	d.Set("resource_group_name", resourceGroup)
	if location := collection.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}
	flattenAndSetTags(d, collection.Tags)

	//resource specific
	if properties := collection.Properties; properties != nil {
		if sku := properties.Sku; sku != nil {
			d.Set("sku", sku.Name)
		}
		d.Set("state", string(properties.State))

		if err := d.Set("quota", flattenAzureArmSchedulerJobCollectionQuota(properties.Quota)); err != nil {
			return fmt.Errorf("Error flattening quota for Job Collection %q (Resource Group %q): %+v", *collection.Name, resourceGroup, err)
		}
	}

	return nil
}

func resourceArmSchedulerJobCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).schedulerJobCollectionsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["jobCollections"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Deleting Scheduler Job Collection %q (resource group %q)", name, resourceGroup)

	future, err := client.Delete(ctx, resourceGroup, name)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error issuing delete request for Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()
	err = future.WaitForCompletionRef(waitCtx, client.Client)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error waiting for deletion of Scheduler Job Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	return nil
}

func expandAzureArmSchedulerJobCollectionQuota(d *schema.ResourceData) *scheduler.JobCollectionQuota {
	if qb, ok := d.Get("quota").([]interface{}); ok && len(qb) > 0 {
		quota := scheduler.JobCollectionQuota{
			MaxRecurrence: &scheduler.JobMaxRecurrence{},
		}

		quotaBlock := qb[0].(map[string]interface{})

		if v, ok := quotaBlock["max_job_count"].(int); ok {
			quota.MaxJobCount = utils.Int32(int32(v))
		}
		if v, ok := quotaBlock["max_recurrence_frequency"].(string); ok {
			quota.MaxRecurrence.Frequency = scheduler.RecurrenceFrequency(v)
		}
		if v, ok := quotaBlock["max_recurrence_interval"].(int); ok && v > 0 {
			quota.MaxRecurrence.Interval = utils.Int32(int32(v))
		} else if v, ok := quotaBlock["max_retry_interval"].(int); ok && v > 0 { //todo remove once max_retry_interval is removed
			quota.MaxRecurrence.Interval = utils.Int32(int32(v))
		}

		return &quota
	}

	return nil
}

func flattenAzureArmSchedulerJobCollectionQuota(quota *scheduler.JobCollectionQuota) []interface{} {

	if quota == nil {
		return nil
	}

	quotaBlock := make(map[string]interface{})

	if v := quota.MaxJobCount; v != nil {
		quotaBlock["max_job_count"] = *v
	}
	if recurrence := quota.MaxRecurrence; recurrence != nil {
		if v := recurrence.Interval; v != nil {
			quotaBlock["max_recurrence_interval"] = *v
			quotaBlock["max_retry_interval"] = *v //todo remove once max_retry_interval is retired
		}

		quotaBlock["max_recurrence_frequency"] = string(recurrence.Frequency)
	}

	return []interface{}{quotaBlock}
}

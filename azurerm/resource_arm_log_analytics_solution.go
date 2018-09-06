package azurerm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/operationsmanagement/mgmt/2015-11-01-preview/operationsmanagement"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmLogAnalyticsSolution() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmLogAnalyticsSolutionCreateUpdate,
		Read:   resourceArmLogAnalyticsSolutionRead,
		Update: resourceArmLogAnalyticsSolutionCreateUpdate,
		Delete: resourceArmLogAnalyticsSolutionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 30),
			Update: schema.DefaultTimeout(time.Minute * 30),
			Delete: schema.DefaultTimeout(time.Minute * 30),
		},

		Schema: map[string]*schema.Schema{
			"solution_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameDiffSuppressSchema(),

			"plan": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"publisher": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"promotion_code": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"product": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"workspace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_resource_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},
		},
	}
}

func resourceArmLogAnalyticsSolutionCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).solutionsClient
	ctx := meta.(*ArmClient).StopContext
	log.Printf("[INFO] preparing arguments for AzureRM Log Analytics solution creation.")

	// The resource requires both .name and .plan.name are set in the format
	// "SolutionName(WorkspaceName)". Feedback will be submitted to the OMS team as IMO this isn't ideal.
	solutionName := d.Get("solution_name").(string)
	workspaceName := d.Get("workspace_name").(string)
	name := fmt.Sprintf("%s(%s)", solutionName, workspaceName)
	resGroup := d.Get("resource_group_name").(string)

	if d.IsNewResource() {
		// first check if there's one in this subscription requiring import
		resp, err := client.Get(ctx, resGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Error checking for the existence of Log Analytics Solution %q (Workspace %q / Resource Group %q): %+v", solutionName, workspaceName, resGroup, err)
			}
		}

		if resp.ID != nil {
			return tf.ImportAsExistsError("azurerm_log_analytics_solution", *resp.ID)
		}
	}

	solutionPlan := expandAzureRmLogAnalyticsSolutionPlan(d)
	solutionPlan.Name = &name

	location := azureRMNormalizeLocation(d.Get("location").(string))
	workspaceID := d.Get("workspace_resource_id").(string)

	parameters := operationsmanagement.Solution{
		Name:     &name,
		Location: &location,
		Plan:     &solutionPlan,
		Properties: &operationsmanagement.SolutionProperties{
			WorkspaceResourceID: &workspaceID,
		},
	}

	waitCtx, cancel := context.WithTimeout(ctx, d.Timeout(tf.TimeoutForCreateUpdate(d)))
	defer cancel()
	res, err := client.CreateOrUpdate(waitCtx, resGroup, name, parameters)
	// Currently this is required to work around successful creation resulting in an error being returned
	if err != nil && res.Response().StatusCode != 201 {
		if resp := res.Response(); resp != nil {
			if resp.StatusCode != 201 {
				return err
			}
		}
	}

	solution, err := client.Get(ctx, resGroup, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Log Analytics Solution %q (Resource Group %q): %+v", name, resGroup, err)
	}

	if solution.ID == nil {
		return fmt.Errorf("Cannot read Log Analytics Solution %q (Resource Group %q) ID", name, resGroup)
	}

	d.SetId(*solution.ID)

	return resourceArmLogAnalyticsSolutionRead(d, meta)

}

func resourceArmLogAnalyticsSolutionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).solutionsClient
	ctx := meta.(*ArmClient).StopContext
	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["solutions"]

	resp, err := client.Get(ctx, resGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on AzureRM Log Analytics solutions '%s': %+v", name, err)
	}

	if resp.Plan == nil {
		return fmt.Errorf("Error making Read request on AzureRM Log Analytics solutions '%s': Plan was nil", name)
	}

	d.Set("resource_group_name", resGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}

	// Reversing the mapping used to get .solution_name
	// expecting resp.Name to be in format "SolutionName(WorkspaceName)".
	if resp.Name != nil && strings.Contains(*resp.Name, "(") {
		if parts := strings.Split(*resp.Name, "("); len(parts) == 2 {
			d.Set("solution_name", parts[0])
			workspaceName := strings.TrimPrefix(parts[1], "(")
			workspaceName = strings.TrimSuffix(workspaceName, ")")
			d.Set("workspace_name", workspaceName)
		} else {
			return fmt.Errorf("Error making Read request on AzureRM Log Analytics solutions '%v': isn't in expected format 'Solution(WorkspaceName)'", resp.Name)
		}
	} else {
		return fmt.Errorf("Error making Read request on AzureRM Log Analytics solutions '%v': isn't in expected format 'Solution(WorkspaceName)'", resp.Name)
	}

	if props := resp.Properties; props != nil {
		d.Set("workspace_resource_id", props.WorkspaceResourceID)
	}
	if plan := resp.Plan; plan != nil {
		d.Set("plan", flattenAzureRmLogAnalyticsSolutionPlan(*resp.Plan))
	}
	return nil
}

func resourceArmLogAnalyticsSolutionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).solutionsClient
	ctx := meta.(*ArmClient).StopContext
	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["solutions"]

	future, err := client.Delete(ctx, resGroup, name)
	if err != nil {
		return fmt.Errorf("Error deleting Log Analytics Solution %q (Resource Group %q): %+v", name, resGroup, err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()
	err = future.WaitForCompletionRef(waitCtx, client.Client)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}

		return fmt.Errorf("Error waiting for deletion of Log Analytics Solution %q (Resource Group %q): %+v", name, resGroup, err)
	}

	return nil
}

func expandAzureRmLogAnalyticsSolutionPlan(d *schema.ResourceData) operationsmanagement.SolutionPlan {
	plans := d.Get("plan").([]interface{})
	plan := plans[0].(map[string]interface{})

	name := plan["name"].(string)
	publisher := plan["publisher"].(string)
	promotionCode := plan["promotion_code"].(string)
	product := plan["product"].(string)

	expandedPlan := operationsmanagement.SolutionPlan{
		Name:          utils.String(name),
		PromotionCode: utils.String(promotionCode),
		Publisher:     utils.String(publisher),
		Product:       utils.String(product),
	}

	return expandedPlan
}

func flattenAzureRmLogAnalyticsSolutionPlan(plan operationsmanagement.SolutionPlan) []interface{} {
	plans := make([]interface{}, 0)
	values := make(map[string]interface{})

	values["name"] = *plan.Name
	values["product"] = *plan.Product
	values["promotion_code"] = *plan.PromotionCode
	values["publisher"] = *plan.Publisher

	return append(plans, values)
}

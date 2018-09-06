package azurerm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAzureRMLogicAppTriggerRecurrence_month(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Month", 1),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Month"),
					resource.TestCheckResourceAttr(resourceName, "interval", "1"),
				),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_requiresImport(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Month", 1),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
				),
			},
			{
				Config:      testAccAzureRMLogicAppTriggerRecurrence_requiresImport(ri, location, "Month", 1),
				ExpectError: testRequiresImportError("azurerm_logic_app_trigger_recurrence"),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_week(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Week", 2),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Week"),
					resource.TestCheckResourceAttr(resourceName, "interval", "2"),
				),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_day(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Day", 3),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Day"),
					resource.TestCheckResourceAttr(resourceName, "interval", "3"),
				),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_minute(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Minute", 4),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Minute"),
					resource.TestCheckResourceAttr(resourceName, "interval", "4"),
				),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_second(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Second", 30),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Second"),
					resource.TestCheckResourceAttr(resourceName, "interval", "30"),
				),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_hour(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Hour", 4),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Hour"),
					resource.TestCheckResourceAttr(resourceName, "interval", "4"),
				),
			},
		},
	})
}

func TestAccAzureRMLogicAppTriggerRecurrence_update(t *testing.T) {
	resourceName := "azurerm_logic_app_trigger_recurrence.test"
	ri := acctest.RandInt()
	location := testLocation()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLogicAppWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Month", 1),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Month"),
					resource.TestCheckResourceAttr(resourceName, "interval", "1"),
				),
			},
			{
				Config: testAccAzureRMLogicAppTriggerRecurrence_basic(ri, location, "Month", 3),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLogicAppTriggerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "Month"),
					resource.TestCheckResourceAttr(resourceName, "interval", "3"),
				),
			},
		},
	})
}

func testAccAzureRMLogicAppTriggerRecurrence_basic(rInt int, location, frequency string, interval int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_logic_app_workflow" "test" {
  name = "acctestlaw-%d"
  location = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_logic_app_trigger_recurrence" "test" {
  name         = "frequency-trigger"
  logic_app_id = "${azurerm_logic_app_workflow.test.id}"
  frequency    = "%s"
  interval     = %d
}
`, rInt, location, rInt, frequency, interval)
}

func testAccAzureRMLogicAppTriggerRecurrence_requiresImport(rInt int, location, frequency string, interval int) string {
	template := testAccAzureRMLogicAppTriggerRecurrence_basic(rInt, location, frequency, interval)
	return fmt.Sprintf(`
%s

resource "azurerm_logic_app_trigger_recurrence" "import" {
  name         = "${azurerm_logic_app_trigger_recurrence.test.name}"
  logic_app_id = "${azurerm_logic_app_trigger_recurrence.test.logic_app_id}"
  frequency    = "${azurerm_logic_app_trigger_recurrence.test.frequency}"
  interval     = "${azurerm_logic_app_trigger_recurrence.test.interval}"
}
`, template)
}

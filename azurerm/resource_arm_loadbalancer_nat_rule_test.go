package azurerm

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-04-01/network"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMLoadBalancerNatRule_basic(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	natRuleId := fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/acctestRG-%d/providers/Microsoft.Network/loadBalancers/arm-test-loadbalancer-%d/inboundNatRules/%s",
		subscriptionID, ri, ri, natRuleName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
					resource.TestCheckResourceAttr(
						"azurerm_lb_nat_rule.test", "id", natRuleId),
				),
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_requiresImport(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	location := testLocation()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
				),
			},
			{
				Config:      testAccAzureRMLoadBalancerNatRule_requiresImport(ri, natRuleName, location),
				ExpectError: testRequiresImportError("azurerm_lb_nat_rule"),
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_removal(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
			{
				Config: testAccAzureRMLoadBalancerNatRule_removal(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleNotExists(natRuleName, &lb),
				),
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_update(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)
	natRule2Name := fmt.Sprintf("NatRule-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_multipleRules(ri, natRuleName, natRule2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRule2Name, &lb),
					resource.TestCheckResourceAttr("azurerm_lb_nat_rule.test2", "frontend_port", "3390"),
					resource.TestCheckResourceAttr("azurerm_lb_nat_rule.test2", "backend_port", "3390"),
				),
			},
			{
				Config: testAccAzureRMLoadBalancerNatRule_multipleRulesUpdate(ri, natRuleName, natRule2Name, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRule2Name, &lb),
					resource.TestCheckResourceAttr("azurerm_lb_nat_rule.test2", "frontend_port", "3391"),
					resource.TestCheckResourceAttr("azurerm_lb_nat_rule.test2", "backend_port", "3391"),
				),
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_reapply(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	deleteNatRuleState := func(s *terraform.State) error {
		return s.Remove("azurerm_lb_nat_rule.test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
					deleteNatRuleState,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_disappears(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
					testCheckAzureRMLoadBalancerNatRuleDisappears(natRuleName, &lb),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_enableFloatingIP(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)
	location := testLocation()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_enableFloatingIP(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
		},
	})
}

func TestAccAzureRMLoadBalancerNatRule_disableFloatingIP(t *testing.T) {
	var lb network.LoadBalancer
	ri := acctest.RandInt()
	natRuleName := fmt.Sprintf("NatRule-%d", ri)
	location := testLocation()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
			{
				Config: testAccAzureRMLoadBalancerNatRule_enableFloatingIP(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
			{
				Config: testAccAzureRMLoadBalancerNatRule_basic(ri, natRuleName, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMLoadBalancerExists("azurerm_lb.test", &lb),
					testCheckAzureRMLoadBalancerNatRuleExists(natRuleName, &lb),
				),
			},
		},
	})
}

func testCheckAzureRMLoadBalancerNatRuleExists(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, exists := findAzureRMLoadBalancerNatRuleByName(lb, natRuleName)
		if !exists {
			return fmt.Errorf("A NAT Rule with name %q cannot be found.", natRuleName)
		}

		return nil
	}
}

func testCheckAzureRMLoadBalancerNatRuleNotExists(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, exists := findAzureRMLoadBalancerNatRuleByName(lb, natRuleName)
		if exists {
			return fmt.Errorf("A NAT Rule with name %q has been found.", natRuleName)
		}

		return nil
	}
}

func testCheckAzureRMLoadBalancerNatRuleDisappears(natRuleName string, lb *network.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*ArmClient).loadBalancerClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		i, exists := findAzureRMLoadBalancerNatRuleByName(lb, natRuleName)
		if !exists {
			return fmt.Errorf("A Nat Rule with name %q cannot be found.", natRuleName)
		}

		currentRules := *lb.LoadBalancerPropertiesFormat.InboundNatRules
		rules := append(currentRules[:i], currentRules[i+1:]...)
		lb.LoadBalancerPropertiesFormat.InboundNatRules = &rules

		id, err := parseAzureResourceID(*lb.ID)
		if err != nil {
			return err
		}

		future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, *lb.Name, *lb)
		if err != nil {
			return fmt.Errorf("Error Creating/Updating Load Balancer %+v", err)
		}

		err = future.WaitForCompletionRef(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Error waiting for the completion of Load Balancer %q (Resource Group %q): %+v", *lb.Name, id.ResourceGroup, err)
		}

		_, err = client.Get(ctx, id.ResourceGroup, *lb.Name, "")
		return err
	}
}

func findAzureRMLoadBalancerNatRuleByName(lb *network.LoadBalancer, name string) (int, bool) {
	if lb == nil || lb.LoadBalancerPropertiesFormat == nil || lb.LoadBalancerPropertiesFormat.InboundNatRules == nil {
		return -1, false
	}

	for i, nr := range *lb.LoadBalancerPropertiesFormat.InboundNatRules {
		if nr.Name != nil && *nr.Name == name {
			return i, true
		}
	}

	return -1, false
}

func testAccAzureRMLoadBalancerNatRule_basic(rInt int, natRuleName string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurerm_resource_group.test.location}"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurerm_public_ip.test.id}"
  }
}

resource "azurerm_lb_nat_rule" "test" {
  location                       = "${azurerm_resource_group.test.location}"
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt)
}

func testAccAzureRMLoadBalancerNatRule_requiresImport(rInt int, natRuleName string, location string) string {
	template := testAccAzureRMLoadBalancerNatRule_basic(rInt, natRuleName, location)
	return fmt.Sprintf(`
%s

resource "azurerm_lb_nat_rule" "import" {
  name                           = "${azurerm_lb_nat_rule.test.name}"
  location                       = "${azurerm_lb_nat_rule.test.location}"
  resource_group_name            = "${azurerm_lb_nat_rule.test.resource_group_name}"
  loadbalancer_id                = "${azurerm_lb_nat_rule.test.loadbalancer_id}"
  protocol                       = "${azurerm_lb_nat_rule.test.protocol}"
  frontend_port                  = "${azurerm_lb_nat_rule.test.frontend_port}"
  backend_port                   = "${azurerm_lb_nat_rule.test.backend_port}"
  frontend_ip_configuration_name = "${azurerm_lb_nat_rule.test.frontend_ip_configuration_name}"
}
`, template)
}

func testAccAzureRMLoadBalancerNatRule_removal(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurerm_resource_group.test.location}"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurerm_public_ip.test.id}"
  }
}
`, rInt, location, rInt, rInt, rInt)
}

func testAccAzureRMLoadBalancerNatRule_multipleRules(rInt int, natRuleName, natRule2Name string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurerm_resource_group.test.location}"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurerm_public_ip.test.id}"
  }
}

resource "azurerm_lb_nat_rule" "test" {
  location                       = "${azurerm_resource_group.test.location}"
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurerm_lb_nat_rule" "test2" {
  location                       = "${azurerm_resource_group.test.location}"
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3390
  backend_port                   = 3390
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt, natRule2Name, rInt)
}

func testAccAzureRMLoadBalancerNatRule_multipleRulesUpdate(rInt int, natRuleName, natRule2Name string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurerm_resource_group.test.location}"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurerm_public_ip.test.id}"
  }
}

resource "azurerm_lb_nat_rule" "test" {
  location                       = "${azurerm_resource_group.test.location}"
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}

resource "azurerm_lb_nat_rule" "test2" {
  location                       = "${azurerm_resource_group.test.location}"
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3391
  backend_port                   = 3391
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt, natRule2Name, rInt)
}

func testAccAzureRMLoadBalancerNatRule_enableFloatingIP(rInt int, natRuleName string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_public_ip" "test" {
  name                         = "test-ip-%d"
  location                     = "${azurerm_resource_group.test.location}"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  public_ip_address_allocation = "static"
}

resource "azurerm_lb" "test" {
  name                = "arm-test-loadbalancer-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  frontend_ip_configuration {
    name                 = "one-%d"
    public_ip_address_id = "${azurerm_public_ip.test.id}"
  }
}

resource "azurerm_lb_nat_rule" "test" {
  location                       = "${azurerm_resource_group.test.location}"
  resource_group_name            = "${azurerm_resource_group.test.name}"
  loadbalancer_id                = "${azurerm_lb.test.id}"
  name                           = "%s"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "one-%d"
}
`, rInt, location, rInt, rInt, rInt, natRuleName, rInt)
}

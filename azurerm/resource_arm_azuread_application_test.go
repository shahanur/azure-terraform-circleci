package azurerm

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMActiveDirectoryApplication_basic(t *testing.T) {
	resourceName := "azurerm_azuread_application.test"
	id := uuid.New().String()
	config := testAccAzureRMActiveDirectoryApplication_basic(id)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMActiveDirectoryApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMActiveDirectoryApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("http://acctest%s", id)),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				),
			},
		},
	})
}

func TestAccAzureRMActiveDirectoryApplication_availableToOtherTenants(t *testing.T) {
	resourceName := "azurerm_azuread_application.test"
	id := uuid.New().String()
	config := testAccAzureRMActiveDirectoryApplication_availableToOtherTenants(id)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMActiveDirectoryApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMActiveDirectoryApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "available_to_other_tenants", "true"),
				),
			},
		},
	})
}

func TestAccAzureRMActiveDirectoryApplication_complete(t *testing.T) {
	resourceName := "azurerm_azuread_application.test"
	id := uuid.New().String()
	config := testAccAzureRMActiveDirectoryApplication_complete(id)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMActiveDirectoryApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMActiveDirectoryApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("http://homepage-%s", id)),
					resource.TestCheckResourceAttr(resourceName, "identifier_uris.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reply_urls.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				),
			},
		},
	})
}

func TestAccAzureRMActiveDirectoryApplication_update(t *testing.T) {
	resourceName := "azurerm_azuread_application.test"
	id := uuid.New().String()
	config := testAccAzureRMActiveDirectoryApplication_basic(id)

	updatedId := uuid.New().String()
	updatedConfig := testAccAzureRMActiveDirectoryApplication_complete(updatedId)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMActiveDirectoryApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMActiveDirectoryApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("http://acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "identifier_uris.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reply_urls.#", "0"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMActiveDirectoryApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", updatedId)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("http://homepage-%s", updatedId)),
					resource.TestCheckResourceAttr(resourceName, "identifier_uris.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reply_urls.#", "1"),
				),
			},
		},
	})
}

func testCheckAzureRMActiveDirectoryApplicationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).applicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Azure AD Application %q does not exist", rs.Primary.ID)
			}
			return fmt.Errorf("Bad: Get on Azure AD applicationsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMActiveDirectoryApplicationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_azuread_application" {
			continue
		}

		client := testAccProvider.Meta().(*ArmClient).applicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Azure AD Application still exists:\n%#v", resp)
	}

	return nil
}

func testAccAzureRMActiveDirectoryApplication_basic(id string) string {
	return fmt.Sprintf(`
resource "azurerm_azuread_application" "test" {
  name = "acctest%s"
}
`, id)
}

func testAccAzureRMActiveDirectoryApplication_requiresImport(id string) string {
	template := testAccAzureRMActiveDirectoryApplication_basic(id)
	return fmt.Sprintf(`
%s

resource "azurerm_azuread_application" "import" {
  name = "${azurerm_azuread_application.test.name}"
}
`, template)
}

func testAccAzureRMActiveDirectoryApplication_availableToOtherTenants(id string) string {
	return fmt.Sprintf(`
resource "azurerm_azuread_application" "test" {
  name                       = "acctest%s"
  identifier_uris            = ["http://%s.hashicorptest.com"]
  available_to_other_tenants = true
}
`, id, id)
}

func testAccAzureRMActiveDirectoryApplication_complete(id string) string {
	return fmt.Sprintf(`
resource "azurerm_azuread_application" "test" {
  name                       = "acctest%s"
  homepage                   = "http://homepage-%s"
  identifier_uris            = ["http://%s.hashicorptest.com"]
  reply_urls                 = ["http://replyurl-%s"]
  oauth2_allow_implicit_flow = true
}
`, id, id, id, id)
}

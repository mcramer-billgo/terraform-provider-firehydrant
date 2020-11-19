package examples

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/firehydrant/terraform-provider-firehydrant/firehydrant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRunbooks(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testFireHydrantIsSetup(t) },
		ProviderFactories: defaultProviderFactories(),
		CheckDestroy:      testRunbookDoesNotExist("firehydrant_runbook.default-incident-process"),
		Steps: []resource.TestStep{
			{
				Config: testRunbookResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("firehydrant_runbook.default-incident-process", "name", rName),
					resource.TestCheckResourceAttr("firehydrant_runbook.default-incident-process", "description", "this is my description"),
				),
			},
			{
				Config: testRunbookResourceConfig(rName + " updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("firehydrant_runbook.default-incident-process", "name", rName+" updated"),
				),
			},
		},
	})
}

func TestAccRunbookActions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testFireHydrantIsSetup(t) },
		ProviderFactories: defaultProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testRunbookActionSlackChannel,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.firehydrant_runbook_action.create-incident-channel", "name", "Create Incident Channel"),
				),
			},
		},
	})
}

const testRunbookActionSlackChannel = `
data "firehydrant_runbook_action" "create-incident-channel" {
	slug = "create_incident_channel"
	integration_slug = "slack"
	type = "incident"
}
`

const testRunbookDataConfigTpl = `
data "firehydrant_runbook" "default-runbook" {
	id = "%s"
}
`

func testRunbookDataConfig(rName string) string {
	return fmt.Sprintf(testRunbookDataConfigTpl, rName)
}

const testRunbookResourceConfigTpl = `
resource "firehydrant_runbook" "default-incident-process" {
	name = "%s"
	type = "incident"
	description = "this is my description"
}
`

func testRunbookResourceConfig(rName string) string {
	return fmt.Sprintf(testRunbookResourceConfigTpl, rName)
}

func testRunbookDoesNotExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[resourceName]

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID was not set")
		}

		c, err := firehydrant.NewRestClient(os.Getenv("FIREHYDRANT_API_KEY"))
		if err != nil {
			return err
		}

		// TODO: Archives dont hide teams from the details endpoint
		svc, err := c.Runbooks().Get(context.TODO(), rs.Primary.ID)
		if svc != nil {
			return fmt.Errorf("The runbook existed, when it should not")
		}

		if _, isNotFound := err.(firehydrant.NotFound); !isNotFound {
			return err
		}

		return nil
	}
}

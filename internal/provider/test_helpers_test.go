package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccOrUnitTest(t *testing.T, steps []resource.TestStep) {
	testCase := resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps:                    steps,
	}

	if os.Getenv("TF_ACC") == "" {
		for i := range steps {
			steps[i].PlanOnly = true
			steps[i].ExpectNonEmptyPlan = true
		}
		resource.UnitTest(t, testCase)
		return
	}

	resource.Test(t, testCase)
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Setenv("MINECRAFT_SKIP_CONNECT", "true")
		return
	}

	t.Setenv("MINECRAFT_SKIP_CONNECT", "")

	// address := os.Getenv("TF_VAR_address")
	// if address == "" {
	// 	address = os.Getenv("MINECRAFT_ADDRESS")
	// }
	// if address == "" {
	// 	t.Fatalf("MINECRAFT_ADDRESS or TF_VAR_address must be set when TF_ACC is enabled")
	// }
	// if os.Getenv("TF_VAR_address") == "" {
	// 	t.Setenv("TF_VAR_address", address)
	// }

	// password := os.Getenv("TF_VAR_password")
	// if password == "" {
	// 	password = os.Getenv("MINECRAFT_PASSWORD")
	// }
	// if password == "" {
	// 	t.Fatalf("MINECRAFT_PASSWORD or TF_VAR_password must be set when TF_ACC is enabled")
	// }
	// if os.Getenv("TF_VAR_password") == "" {
	// 	t.Setenv("TF_VAR_password", password)
	// }
}

func testProviderConfig() string {
	return `
variable "address" {
  type    = string
  default = "localhost:27015"
}

variable "password" {
  type    = string
  default = ""
}

provider "minecraft" {
  address  = var.address
  password = var.password
}
`
}

func testBlockPositionAttributes(prefix string, x, y, z int) string {
	return fmt.Sprintf(`
  %s = {
    x = %d
    y = %d
    z = %d
  }
`, prefix, x, y, z)
}
